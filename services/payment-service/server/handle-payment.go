package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/paymentintent"
	"github.com/stripe/stripe-go/paymentmethod"
	"github.com/veganbase/backend/chassis"
	_ "github.com/veganbase/backend/services/payment-service/db"
	"github.com/veganbase/backend/services/payment-service/events"
	"github.com/veganbase/backend/services/payment-service/model"
	pur "github.com/veganbase/backend/services/purchase-service/model"
)

func (s *Server) createPaymentIntentWithAuth(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	authInfo := chassis.AuthInfoFromContext(r.Context())
	if authInfo.AuthMethod == chassis.NoAuth {
		return chassis.NotFound(w)
	}

	//TODO: may be a good idea to check if the user is the owner of the purchase
	return s.createPaymentIntent(w, r)
}

// get all information of a purchase and process the payment
func (s *Server) createPaymentIntent(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	purchase := pur.Purchase{}
	body, err := chassis.ReadBody(r, 0)
	if err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	if err = json.Unmarshal(body, &purchase); err != nil {
		return chassis.BadRequest(w, err.Error())
	}

	totalAmount := make(map[string]int)

	//summing up all values with the same currency
	//different currencies will be charged separately
	for _, item := range purchase.Items {
		totalAmount[item.Currency] += item.Price * item.Quantity
	}

	//summing up delivery fees with the items total
	for _, fee := range *purchase.DeliveryFees {
		totalAmount[fee.Currency] += fee.Price
	}

	var isSimplePurchase bool
	if purchase.PaymentMethod != "" {
		isSimplePurchase = true
	}

	var customerID string
	if !isSimplePurchase {
		paymentMethod, err := s.userSvc.GetDefaultPaymentMethod(purchase.BuyerID)
		if err != nil {
			return nil, errors.New("error getting the user default payment method")
		}
		purchase.PaymentMethod = paymentMethod.PaymentMethodID

		//getting customer on the database
		customer, err := s.userSvc.GetCustomer(purchase.BuyerID)
		if err != nil {
			return nil, errors.New("error getting the user customer reference")
		}
		customerID = customer.CustomerID
	}

	//validate payment method using Stripe API
	if _, err = paymentmethod.Get(purchase.PaymentMethod, nil); err != nil {
		return nil, errors.New("error validating your payment method within Stripe" + err.Error())
	}

	var descriptor *string
	if purchase.Site != nil {
		siteMap := s.siteSvc.Sites()
		for _, site := range siteMap {
			if site.URL == *purchase.Site {
				descriptor = stripe.String(site.Name)
			}
		}
	}

	var errorOccurred bool
	intents := []model.PaymentIntent{}
	for currency, amount := range totalAmount {
		params := &stripe.PaymentIntentParams{
			PaymentMethod:       stripe.String(purchase.PaymentMethod),
			Amount:              stripe.Int64(int64(amount)),
			Currency:            stripe.String(strings.ToLower(currency)),
			Confirm:             stripe.Bool(true),
			StatementDescriptor: descriptor,
			//TransferGroup: stripe.String(purchase.Id),
		}
		if !isSimplePurchase {
			params.Customer = stripe.String(customerID)
		}

		pi, err := paymentintent.New(params)
		if err != nil {
			//casting a generic error to stripe.Error
			if stripeErr, ok := err.(*stripe.Error); ok {
				if cardErr, ok := stripeErr.Err.(*stripe.CardError); ok && stripeErr.PaymentIntent != nil {
					payInt := model.PaymentIntent{
						StripeIntentId: stripeErr.PaymentIntent.ID,
						Origin:         purchase.Id,
						Status:         string(stripeErr.Code),
						OriginAmount:   stripeErr.PaymentIntent.Amount,
						Currency:       stripeErr.PaymentIntent.Currency,
					}
					if err = s.db.CreatePaymentIntent(&payInt); err != nil {
						j, _ := json.Marshal(payInt)
						s.LogError(payInt.StripeIntentId, "error while saving intent to database: "+string(j))
					}
					//even when failing, we must return the payment attempt
					intents = append(intents, payInt)
					chassis.Emit(s, events.PaymentCreated, payInt.StripeIntentId)

					s.LogError(payInt.StripeIntentId, "card declined with code: "+string(cardErr.DeclineCode))
				}
			}
			errorOccurred = true
			break
		}

		//adding successful intent to be saved
		payInt := model.PaymentIntent{
			StripeIntentId: pi.ID,
			Origin:         purchase.Id,
			Status:         string(pi.Status),
			OriginAmount:   pi.Amount,
			Currency:       pi.Currency,
		}

		if err = s.db.CreatePaymentIntent(&payInt); err != nil {
			j, _ := json.Marshal(payInt)
			s.LogError(payInt.StripeIntentId, "error while saving payment intent:"+string(j))
			errorOccurred = true
			break
		}
		chassis.Emit(s, events.PaymentCreated, payInt.StripeIntentId)

		//verifying if 3D secure auth is required
		if pi.Status == stripe.PaymentIntentStatusRequiresAction && pi.NextAction.Type == "use_stripe_sdk" {
			payInt.RequiresAction = true
			payInt.PaymentIntentClientSecret = &pi.ClientSecret
		}

		intents = append(intents, payInt)

	}

	if errorOccurred {
		s.rollback(intents)
	}

	return intents, err
}

// confirmPaymentIntent confirm the payment on stripe and updates vb database with the latest status
func (s *Server) confirmPaymentIntent(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	stripe.Key = s.stripeKey

	paymentIntentId := chi.URLParam(r, "pi_id")

	stripeIntent, err := paymentintent.Confirm(paymentIntentId, nil)
	if err != nil {
		return nil, err
	}

	vbIntent, err := s.db.PaymentIntentByIntentId(paymentIntentId)
	if err != nil {
		return nil, err
	}

	if vbIntent.Status != string(stripeIntent.Status) {
		vbIntent.Status = string(stripeIntent.Status)

		err = s.db.UpdatePaymentIntent(vbIntent)
		if err != nil {
			return nil, err
		}
		chassis.Emit(s, events.PaymentUpdated, paymentIntentId)
		return vbIntent, err
	}

	return nil, errors.New("payment intent status unchanged")
}

// rollback a set of payment intents by canceling them on stripe and saving the status on database
// useful when we couldn't fulfill all payments of a certain purchase (one of them fails, in case of multiple payments)
func (s *Server) rollback(intents []model.PaymentIntent) {
	for _, intent := range intents {
		if err := s.CancelIntent(intent.StripeIntentId); err != nil {
			s.LogError(intent.StripeIntentId, "error occurred canceling intent: "+err.Error())
			return
		}
		paymentIntent, _ := s.RetrieveIntent(intent.StripeIntentId)
		if paymentIntent.Status == "canceled" {
			intent.Status = "canceled"
			if err := s.db.UpdatePaymentIntent(&intent); err != nil {
				s.LogError(intent.StripeIntentId, "error occurred while updating payment status: "+err.Error())
				return
			}
		}
	}
	return
}
