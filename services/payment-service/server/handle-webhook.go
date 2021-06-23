package server

import (
	"encoding/json"
	"fmt"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/paymentintent"
	"github.com/stripe/stripe-go/transfer"
	"github.com/stripe/stripe-go/webhook"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/payment-service/db"
	"github.com/veganbase/backend/services/payment-service/events"
	"github.com/veganbase/backend/services/payment-service/model"
	pur "github.com/veganbase/backend/services/purchase-service/model"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
)

const (
	DefaultFee  float64 = 0.15
	MinFraction         = 0.00005
)

func (s *Server) webhookHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var event stripe.Event

	if s.webhookSecret != "" {
		//this is required to validate that the one calling the webhook path is Stripe
		event, err = webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), s.webhookSecret)
		if err != nil {
			return chassis.BadRequest(w, "error occurred while validating Stripe's signature")
		}
	} else {
		err := json.Unmarshal(payload, &event)
		if err != nil {
			return nil, err
		}
	}

	objectType := event.Data.Object["object"].(string)
	if check, err := s.db.ReceivedEventByEventId(event.ID); err != nil {
		if err == db.ErrReceivedEventNotFound {
			receivedEvent := model.ReceivedEvent{
				EventId:        event.ID,
				IdempotencyKey: event.Request.IdempotencyKey,
				EventType:      event.Type,
				IsHandled:      false,
			}
			if err = s.db.CreateReceivedEvent(&receivedEvent); err != nil {
				s.LogError(event.ID, "error while saving event on database: "+err.Error())
			}
		} else {
			s.LogError(event.ID, "error checking event on database: "+err.Error())
		}
	} else {
		if check != nil {
			fmt.Printf("🔔  Webhook received, but the event is already in the database and wont be handled: %s\n", event.Type)
			s.LogError(event.ID, "event "+event.ID+" of type '"+event.Type+"' received but is duplicated. Will not be handled.")
			return nil, nil
		}
	}

	var handled bool
	switch objectType {
	case "payment_intent":
		var pi *stripe.PaymentIntent
		if err = json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return nil, err
		}
		handled = true
		go s.HandlePaymentIntent(&event, pi, true)
	case "source":
		var source *stripe.Source
		if err := json.Unmarshal(event.Data.Raw, &source); err != nil {
			return nil, err
		}
		handled = true
		go s.HandleSource(event, source)
	}

	if handled {
		fmt.Printf("🔔  Webhook received and will be handled by other routine! %s\n", event.Type)
		s.UpdateReceivedEvent(event.ID, handled)
	} else {
		fmt.Printf("🔔  Webhook received and not handled! %s\n", event.Type)
	}

	return nil, err
}

func (s *Server) confirmIntent(paymentIntent string, source *stripe.Source) error {
	pi, err := paymentintent.Get(paymentIntent, nil)
	if err != nil {
		return fmt.Errorf("payments: error fetching payment intent for confirmation: %v", err)
	}

	if pi.Status != "requires_payment_method" {
		return fmt.Errorf("payments: PaymentIntent already has a status of %s", pi.Status)
	}

	params := &stripe.PaymentIntentConfirmParams{
		Source: stripe.String(source.ID),
	}
	_, err = paymentintent.Confirm(pi.ID, params)
	if err != nil {
		return fmt.Errorf("payments: error confirming PaymentIntent: %v", err)
	}

	return nil
}

func (s *Server) HandlePaymentIntent(event *stripe.Event, pi *stripe.PaymentIntent, isFirstAttempt bool) {

	//getting payment intent on database
	payInt, err := s.db.PaymentIntentByIntentId(pi.ID)
	if err != nil {
		if isFirstAttempt {
			pe := model.PendingEvent{EventID: event.ID, IntentID: pi.ID, Reason: err.Error(), Attempts: 0,}
			if err = s.db.CreatePendingEvent(&pe); err != nil {
				s.LogError(event.ID, "Error while creating a pending event with event_id="+
					pe.EventID+" and intent_id="+pe.IntentID)
			}
		}
		return
	}

	if !isFirstAttempt {
		if err = s.db.DeletePendingEvent(event.ID); err != nil {
			s.LogError(event.ID, "error while deleting already processed event: "+err.Error())
		}
	}

	switch event.Type {
	case "payment_intent.succeeded":
		fmt.Printf("🔔  Webhook received! Payment for PaymentIntent %s succeeded\n", pi.ID)

		if err := s.updateStatus(pi, payInt); err != nil {
			s.LogError(event.ID, "Error while updating status of payment-intent: "+err.Error())
		}

		go s.sendPaymentReceivedNotification(payInt, true)

		message, err := s.validatePayments(payInt)
		if err != nil {
			s.LogError(event.ID, message+err.Error())
		}
	case "payment_intent.payment_failed":
		if pi.LastPaymentError.PaymentMethod != nil {
			fmt.Printf(
				"🔔  Webhook received! Payment on %s %s for PaymentIntent %s failed\n",
				"payment_method",
				pi.LastPaymentError.PaymentMethod.ID,
				pi.ID,
			)
			if err := s.updateStatus(pi, payInt); err != nil {
				s.LogError(event.ID, "Error while updating status of payment-intent: "+err.Error())
			}
			//updating purchase on purchase-service
			if _, err := s.purchaseSvc.UpdatePurchaseStatus(payInt.Origin, "failed"); err != nil {
				s.LogError(event.ID, "Error while updating purchase status on purchase-service: "+err.Error())
			}
			//TODO: EMIT MSG TELLING THE BUYER THAT ONE PAYMENT FAILED
		} else {
			fmt.Printf(
				"🔔  Webhook received! Payment on %s %s for PaymentIntent %s failed\n",
				"source",
				pi.LastPaymentError.Source.ID,
				pi.ID,
			)
			if err := s.updateStatus(pi, payInt); err != nil {
				s.LogError(event.ID, "Error while updating status of payment-intent: "+err.Error())
			}
			//updating purchase on purchase-service
			if _, err := s.purchaseSvc.UpdatePurchaseStatus(payInt.Origin, "failed"); err != nil {
				s.LogError(event.ID, "Error while updating purchase status on purchase-service: "+err.Error())
			}
		}
		go s.sendPaymentReceivedNotification(payInt, false)
	}

}

func (s *Server) HandleSource(event stripe.Event, source *stripe.Source) (bool, error) {
	paymentIntent := source.Metadata["paymentIntent"]
	if paymentIntent == "" {
		return false, nil
	}

	if source.Status == "chargeable" {
		fmt.Printf("🔔  Webhook received! The source %s is chargeable\n", source.ID)
		return true, s.ConfirmIntent(paymentIntent, source)
	} else if source.Status == "failed" || source.Status == "canceled" {
		return true, s.CancelIntent(paymentIntent)
	}

	return false, nil
}

func (s *Server) updateStatus(pi *stripe.PaymentIntent, payInt *model.PaymentIntent) error {
	payInt.Status = string(pi.Status)
	return s.db.UpdatePaymentIntent(payInt)
}

func (s *Server) validatePayments(payInt *model.PaymentIntent) (string, error) {

	//getting all payments with the same purchase as origin that succeeded
	var successPayments *[]model.PaymentIntent
	var err error
	if successPayments, err = s.db.SuccessfulPaymentIntentsByOrigin(payInt.Origin); err != nil {
		return "Error while getting successful payment-intents on database: ", err
	}

	//summing up the absolute value of successful payments (ignoring currency)
	var total int64
	for _, payment := range *successPayments {
		total += payment.OriginAmount
	}

	purchaseInfo, err := s.purchaseSvc.GetPurchaseInfo(payInt.Origin)
	if err != nil {
		return "Error while getting purchase information on purchase-service: ", err
	}

	//summing the absolute prices of items inside the purchase (ignoring currency)
	var expectedValue int
	for _, item := range purchaseInfo.Items {
		expectedValue += item.Price * item.Quantity
	}
	for _, fee := range *purchaseInfo.DeliveryFees {
		expectedValue += fee.Price
	}

	// if the sum of payments equals the expected value, trigger payouts
	if int64(expectedValue) == total {
		return s.triggerPayouts(purchaseInfo, successPayments)
	}

	return "success", nil
}

func (s *Server) triggerPayouts(pur *pur.FullPurchase, successPayments *[]model.PaymentIntent) (string, error) {
	payments := make(map[string]*stripe.PaymentIntent)
	// getting all stripe's payment intent information so we can wire the transfer to it
	// in order to only transfer when the funds are released
	for _, pi := range *successPayments {
		paymentIntent, _ := paymentintent.Get(
			pi.StripeIntentId,
			nil,
		)
		payments[paymentIntent.Currency] = paymentIntent
	}

	//TODO: GET FEE BASED ON URL OF pur.Site
	//  THIS CAN BE O(1)
	var fee float64
	fee = DefaultFee
	siteMap := s.siteSvc.Sites()
	for _, site := range siteMap {
		if site.URL == *pur.Site {
			fee = site.Fee
			break
		}
	}

	//creating transfers related to sellers
	for _, order := range *pur.Orders {
		orderTotal := map[string]int{}
		//summing all items with the same currency
		for _, item := range order.Items {
			orderTotal[strings.ToLower(item.Currency)] += item.Price * item.Quantity
		}

		//adding delivery fee order total
		if order.DeliveryFee != nil {
			if order.DeliveryFee.Currency != "" {
				orderTotal[strings.ToLower(order.DeliveryFee.Currency)] += order.DeliveryFee.Price
			}
		}

		for currency, total := range orderTotal {
			if message, err := s.performTransfer(total, fee, currency, order.Seller, order.Origin, payments); err != nil {
				return message, err
			}
		}

		if _, err := s.purchaseSvc.UpdateOrderPaymentStatus(order.Id, "completed"); err != nil {
			return "Error while setting the order status as 'complete' on purchase-service: ", err
		}
		//notify seller by email
		if msg, err := s.buildOrderEmailMsg(order.Id, order.Seller); err == nil {
			chassis.Emit(s, events.SaleCompleteTopic, msg)
		}
	}

	//creating transfers related to each booking
	for _, booking := range *pur.Bookings {

		bookingTotal := booking.BookingInfo.Price * booking.BookingInfo.Quantity

		if message, err := s.performTransfer(bookingTotal, fee, strings.ToLower(booking.BookingInfo.Currency),
			booking.Host, booking.Origin, payments); err != nil {
			return message, err
		}

		if _, err := s.purchaseSvc.UpdateBookingPaymentStatus(booking.Id, "completed"); err != nil {
			return "Error while setting the booking status as 'complete' on purchase-service: ", err
		}
		//notify host by email
		if msg, err := s.buildOrderEmailMsg(booking.Id, booking.Host); err == nil {
			chassis.Emit(s, events.SaleCompleteTopic, msg)
		}

	}

	//after creating all transfers, we change the purchase status to completed
	if _, err := s.purchaseSvc.UpdatePurchaseStatus(pur.Id, "completed"); err != nil {
		return "Error while updating purchase status on purchase-service: ", err
	}
	//notify buyer by email
	if msg, err := s.buildPaymentEmailMsg(pur.Id, pur.BuyerID, "success"); err == nil {
		chassis.Emit(s, events.PaymentStatusTopic, msg)
	}

	for _, order := range *pur.Orders {
		payload := events.BuildOrderPlacedEventPayload(order)

		if err := chassis.TriggerWebhookEvent(s, order.Seller, events.OrderPlaced, s.livemode, payload); err != nil {
			s.LogError(order.Id, "Error triggering webhook event: "+ string(payload))
		}
	}

	return "success", nil
}

//performTransfer will perform a stripe transfer to the destination account number and all the logging related to it.
// these logs include a transfer row on transfers table and one in transfer_remainders for accounting purposes.
func (s *Server) performTransfer(total int, fee float64, currency, destinationId, origin string, payments map[string]*stripe.PaymentIntent) (string, error) {

	totalAsFloat := float64(total)
	feeToBeCollected := totalAsFloat * fee // this is the fee that should be collected by VB
	//we cannot collect or transfer fraction of cents, so we store the remainders
	feeCollected, feeRemainder := math.Modf(feeToBeCollected)
	totalAfterFee, totalRemainder := math.Modf(totalAsFloat - feeToBeCollected)
	sourceTransaction := stripe.String(payments[currency].Charges.Data[0].ID)
	payoutAccount, err := s.userSvc.GetPayoutAccount(destinationId)
	if err != nil {
		//if an error occurred while getting payout account, we add the transfer to a queue to be processed later
		return s.CreatePendingTransfer(origin, destinationId, currency, *sourceTransaction, total, totalAfterFee, feeCollected, feeRemainder, totalRemainder, err.Error())
	}

	//TODO: CHECK IF ACCOUNT IS ALRIGHT IN STRIPE
	//stripeAccount, err := account.GetByID(
	//	payoutAccount.AccountNumber,
	//	nil,
	//)

	transferParams := &stripe.TransferParams{
		Amount:      stripe.Int64(int64(totalAfterFee)),
		Currency:    stripe.String(currency),
		Destination: stripe.String(payoutAccount.AccountNumber),
		//TransferGroup: stripe.String(order.Id),
		SourceTransaction: sourceTransaction,
	}

	tr, err := transfer.New(transferParams)
	if err != nil {
		return s.CreatePendingTransfer(origin, destinationId, currency, *sourceTransaction, total, totalAfterFee, feeCollected, feeRemainder, totalRemainder, err.Error())
	}

	transRemainder := model.TransferRemainder{
		TransferId:           tr.ID,
		Destination:          destinationId,
		DestinationAccount:   payoutAccount.AccountNumber,
		Currency:             currency,
		TotalValue:           total,
		TransferredValue:     int(totalAfterFee),
		FeeValue:             int(feeCollected),
		FeeRemainder:         feeRemainder,
		TransferredRemainder: totalRemainder,
	}
	if err = s.db.CreateTransfers(&transRemainder, origin); err != nil {
		return "Error while saving a transfer remainder on database: ", err
	}

	return "success", nil
}

func (s *Server) CreatePendingTransfer(origin, destinationId, currency, sourceTransaction string, total int,
	totalAfterFee, feeCollected, feeRemainder, totalRemainder float64, err string) (string, error) {

	pending := model.PendingTransfer{
		Origin:               origin,
		Destination:          destinationId,
		Currency:             currency,
		SourceTransaction:    sourceTransaction,
		TotalValue:           total,
		TransferredValue:     int(totalAfterFee),
		FeeValue:             int(feeCollected),
		FeeRemainder:         feeRemainder,
		TransferredRemainder: totalRemainder,
		Reason:               err,
	}

	if err := s.db.CreatePendingTransfer(&pending); err != nil {
		errorDetails := fmt.Sprintf("(origin=%s, destination=%s, currency=%s, sourceTransaction=%s, totalValue=%d, "+
			"trasferredValue=%d, feeValue=%d, feeRemainder=%f, transferredRemainder=%f, reason=%s",
			origin, destinationId, currency, sourceTransaction, total, int(totalAfterFee), int(feeCollected),
			feeRemainder, totalRemainder, err)

		return "Error while adding a transfer (" + errorDetails + ") to pending queue: ", err
	}
	//queued transfer successfully
	return "success", nil
}
