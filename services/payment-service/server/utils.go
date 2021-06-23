package server

import (
	"fmt"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/paymentintent"
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/payment-service/model"
)

func (s *Server) RetrieveIntent(paymentIntent string) (*stripe.PaymentIntent, error) {
	pi, err := paymentintent.Get(paymentIntent, nil)
	if err != nil {
		return nil, fmt.Errorf("payments: error fetching payment intent: %v", err)
	}

	return pi, nil
}

func (s *Server) ConfirmIntent(paymentIntent string, source *stripe.Source) error {
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
	pi, err = paymentintent.Confirm(pi.ID, params)
	if  err != nil {
		return fmt.Errorf("payments: error confirming PaymentIntent: %v", err)
	}
	//getting payment intent on vb database
	vbIntent, err :=s.db.PaymentIntentByIntentId(pi.ID)
	if err != nil {
		return fmt.Errorf("payments: error while getting vb PaymentIntent: %v", err)
	}

	//updating status
	vbIntent.Status = string(pi.Status)
	return s.db.UpdatePaymentIntent(vbIntent)


}

func (s *Server) CancelIntent(paymentIntent string) error {
	if _, err := paymentintent.Cancel(paymentIntent, nil); err != nil {
		return fmt.Errorf("payments: error canceling PaymentIntent: %v", err)
	}

	//getting payment intent on vb database
	vbIntent, err :=s.db.PaymentIntentByIntentId(paymentIntent)
	if err != nil {
		return fmt.Errorf("payments: error while getting vb PaymentIntent: %v", err)
	}

	//updating status
	vbIntent.Status = "cancelled"
	return s.db.UpdatePaymentIntent(vbIntent)

}

func (s *Server) LogError(eventId, error string) {
	log := model.ErrorLog{
		EventId:   eventId,
		Error:     error,
	}
	if err := s.db.CreateErrorLog(&log); err != nil {
		fmt.Printf("🔔  Error occurred while logging an error on database with event_id = %s, error = %s \n", eventId, error)
	}
}

func (s *Server) UpdateReceivedEvent(eventId string, isHandled bool) {
	event, err := s.db.ReceivedEventByEventId(eventId)
	if err != nil {
		s.LogError(eventId, "Error while getting the received event: "+ err.Error())
		return
	}
	event.IsHandled = isHandled
	if err = s.db.UpdateReceivedEvent(event); err != nil {
		s.LogError(eventId, "Error while updating the received event: "+ err.Error())
	}
	return
}

func (s *Server) buildPaymentEmailMsg(purchaseId, buyerId, status string) (*chassis.GenericEmailMsg, error){
	info, err := s.userSvc.GetNotificationInfo(buyerId)
	if err != nil {
		//LOG ERROR
	}
	msg := chassis.GenericEmailMsg{}
	msg.FixedFields = chassis.FixedFields{
		Email:    info.Email,
		Language: "en",
		Site:     "veganbase",
	}

	other := make(chassis.GenericMap)
	other["customer_name"] = info.Name
	other["purchase_id"] = purchaseId
	other["payment_status"] = status

	msg.Data = other

	return &msg, nil
}

func (s *Server) buildOrderEmailMsg(orderId, sellerId string) (*chassis.GenericEmailMsg, error){
	info, err := s.userSvc.GetNotificationInfo(sellerId)
	if err != nil {
		//LOG ERROR
	}
	msg := chassis.GenericEmailMsg{}
	msg.FixedFields = chassis.FixedFields{
		Email:    info.Email,
		Language: "en",
		Site:     "veganbase",
	}

	other := make(chassis.GenericMap)
	other["dst_name"] = info.Name
	other["sale_id"] = orderId //we use sale_id to generalize orders and bookings on Mailjet's templates
	other["sale_type"] = "order"
	msg.Data = other

	return &msg, nil
}

func (s *Server) buildBookingEmailMsg(bookingId, hostId string) (*chassis.GenericEmailMsg, error){
	info, err := s.userSvc.GetNotificationInfo(hostId)
	if err != nil {
		//LOG ERROR
	}
	msg := chassis.GenericEmailMsg{}
	msg.FixedFields = chassis.FixedFields{
		Email:    info.Email,
		Language: "en",
		Site:     "veganbase",
	}

	other := make(chassis.GenericMap)
	other["dst_name"] = info.Name
	other["sale_id"] = bookingId
	other["sale_type"] = "booking"
	msg.Data = other

	return &msg, nil
}