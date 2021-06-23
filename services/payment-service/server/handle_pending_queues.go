package server

import (
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/event"
	"github.com/stripe/stripe-go/paymentintent"
	"github.com/stripe/stripe-go/transfer"
	"github.com/veganbase/backend/services/payment-service/model"
	"strconv"
	"time"
)

func (s *Server) HandlePendingTransfers() {
	//TODO: SCHEDULE THIS ROUTINE
	period := time.Tick(10 * time.Minute)
	for range period {
		//query all pending transfers
		pendingTransfers, err := s.db.PendingTransfers()
		if err != nil {
			log.Error().Err(err).Msg("obtaining pending transfers on database")
			continue
		}
		for _, pt := range *pendingTransfers {
			payoutAcc, err := s.userSvc.GetPayoutAccount(pt.Destination)
			if err != nil {
				log.Error().Err(err).Msg("obtaining payout account for " + pt.Destination + ".")
				continue
			}
			transferParams := &stripe.TransferParams{
				Amount:            stripe.Int64(int64(pt.TotalValue)),
				Currency:          stripe.String(pt.Currency),
				Destination:       stripe.String(payoutAcc.AccountNumber),
				SourceTransaction: stripe.String(pt.SourceTransaction),
			}

			tr, err := transfer.New(transferParams)
			if err != nil {
				if stripeErr, ok := err.(*stripe.Error); ok {
					log.Error().Err(errors.New(stripeErr.Msg)).Msg("while wiring a transfer on Stripe for " + pt.Destination)
				} else {
					log.Error().Err(err).Msg("while wiring a transfer on Stripe for " + pt.Destination + ".")
				}
				continue
			}

			transRemainder := model.TransferRemainder{
				TransferId:           tr.ID,
				Destination:          pt.Origin,
				DestinationAccount:   payoutAcc.AccountNumber,
				Currency:             pt.Currency,
				TotalValue:           pt.TotalValue,
				TransferredValue:     pt.TransferredValue,
				FeeValue:             pt.FeeValue,
				FeeRemainder:         pt.FeeRemainder,
				TransferredRemainder: pt.TransferredRemainder,
			}
			if err = s.db.CreateTransfers(&transRemainder, pt.Origin); err != nil {
				log.Error().Err(err).Msg("while creating transfers for " + pt.Destination + ".")
				continue
			}
			if err = s.db.DeletePendingTransfer(pt.ID); err != nil {
				log.Error().Err(err).Msg("while deleting pending transfer with ID= " + strconv.FormatInt(pt.ID, 10) + ".")
				continue
			}
			log.Info().Msg("Transfer successfully made to " + pt.Destination)
		}

	}
}

func (s *Server) HandlePendingEvents() {
	//TODO: SCHEDULE THIS ROUTINE
	period := time.Tick(1 * time.Minute)
	for range period {
		//query all pending transfers
		pendingEvents, err := s.db.PendingEvents()
		if err != nil {
			log.Error().Err(err).Msg("obtaining pending events on database")
			continue
		}
		for _, pe := range *pendingEvents {
			event, err := event.Get(pe.EventID, nil)
			if err != nil {
				if stripeErr, ok := err.(*stripe.Error); ok {
					log.Error().Err(errors.New(stripeErr.Msg)).Msg("while acquiring event from Stripe " + pe.EventID + ".")
				} else {
					log.Error().Err(err).Msg("while acquiring event from Stripe " + pe.EventID + ".")
				}
				continue
			}
			intent, err := paymentintent.Get(pe.IntentID, nil)
			if err != nil {
				if stripeErr, ok := err.(*stripe.Error); ok {
					log.Error().Err(errors.New(stripeErr.Msg)).Msg("while acquiring intent from Stripe " + pe.IntentID + ".")
				} else {
					log.Error().Err(err).Msg("while acquiring intent from Stripe " + pe.IntentID + ".")
				}
				continue
			}
			go s.HandlePaymentIntent(event, intent, false)
			//if err = s.db.DeletePendingEvent(pe.EventID); err != nil {
			//	log.Error().Err(err).Msg("while deleting event " + pe.EventID + ".")
			//	continue
			//}
		}
	}
}
