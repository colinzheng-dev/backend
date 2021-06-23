package server

import (
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/payment-service/events"
	"github.com/veganbase/backend/services/payment-service/model"
	pur "github.com/veganbase/backend/services/purchase-service/model"
	usr "github.com/veganbase/backend/services/user-service/model"
	"strconv"
)

func (s *Server) sendPaymentReceivedNotification(p *model.PaymentIntent, success bool) {
	if p.Status != "canceled" && !success {
		p.Status = "canceled"
	}

	purchaseInfo, err := s.purchaseSvc.GetPurchaseInfo(p.Origin)
	if err != nil {
		log.Error().Err(err).Msg("payment-received: could not obtain purchase information from purchase-service")
	}

	userInfo, err := s.userSvc.Info([]string{purchaseInfo.BuyerID})
	if err != nil {
		log.Error().Err(err).Msg("payment-received: could not obtain users' information from user-service")
		return
	}

	//SENDING payment-received notification to buyer
	notification := buildPaymentReceivedNotificationMsg(p, purchaseInfo, userInfo[purchaseInfo.BuyerID])
	if err = chassis.Emit(s, events.PaymentReceivedTopic, notification); err != nil {
		log.Error().Err(err).Msg("payment-received: could not send event")
	}

	return
}


func buildPaymentReceivedNotificationMsg(p *model.PaymentIntent, purchase *pur.FullPurchase, info *usr.Info ) *chassis.GenericEmailMsg {
	data := chassis.GenericMap{}

	data["payment_status"] = p.Status
	data["purchase_id"] = p.Origin
	data["payment_number"] = p.StripeIntentId
	data["customer_name"] = info.Name

	amount := int(p.OriginAmount)
	data["payment_amount"] = strconv.Itoa(amount)
	data["payment_currency"] = p.Currency
	data["payment_formatted_value"] = chassis.FormatCurrencyValue(p.Currency, amount)

	msg := chassis.GenericEmailMsg{
		FixedFields: chassis.FixedFields{
			Site:     *purchase.Site,
			Language: "en",
			Email:    *info.Email,
		},
		Data: data,
	}

	return &msg
}