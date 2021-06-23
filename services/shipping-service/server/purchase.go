package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis/pubsub"
	it "github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/purchase-service/events"
	models "github.com/veganbase/backend/services/purchase-service/model"
	"github.com/veganbase/backend/services/shipping-service/integrations/shippypro/model"
)

func (s *Server) HandleCompletePurchases() {
	newPurchaseCH, _, err := s.PubSub.Subscribe(events.PurchaseCreated, s.AppName, pubsub.Fanout)
	if err != nil {
		log.Fatal().Err(err).
			Msg("unable to subscribe to puchase created topic")
	}
	for {
		purchaseJSON := <-newPurchaseCH
		purchase := models.Purchase{}
		if err := json.Unmarshal(purchaseJSON, &purchase); err != nil {
			log.Error().Err(err).
				Msg("decoding purchase message")
			continue
		}

		//TODO: we should not get the default address. we need the request address
		address, err := s.userSvc.GetDefaultAddress(purchase.BuyerID)
		if err != nil {
			log.Error().Err(err).
				Msg("gettings default address")
			continue
		}

		total := 0.0
		count := int64(0)
		shipmentFee := 0.0
		note := "empty note"
		items := make([]model.Item, 0)
		for _, i := range purchase.Items {
			name := i.OtherInfo["name"]
			if name.(string) == "Shipping fee" {
				shipmentFee = float64(i.Price / 100)
				continue
			}
			itemID := i.ItemId
			if i.ProductType == it.ProductOfferingItem.String() {
				var ok bool
				itemID, ok = i.OtherInfo["original_item_id"].(string)
				if !ok {
					log.Err(fmt.Errorf("failed to get original item id")).
						Msg("productoffering without original id")
				}
			}
			itemInfo, err := s.itemSvc.ItemInfo(itemID)
			if err != nil {
				log.Err(err).Msg("getting item info")
				continue
			}
			img := itemInfo.Pictures[0]

			total += float64(i.Quantity*i.Price) / 100.0
			count += int64(i.Quantity)
			items = append(items, model.Item{
				Title:    name.(string),
				Imageurl: img,
				Quantity: int64(i.Quantity),
				Price:    float64(i.Price) / 100,
				Sku:      i.ItemId,
			})
		}

		client := &http.Client{}

		bodyStruct := model.Order{
			Method: model.PutOrder,
			Params: model.Params{
				ToAddress: model.ToAddress{
					Name:    address.Recipient.FirstName + " " + address.Recipient.LastName,
					Company: address.Recipient.Company,
					Street1: address.StreetAddress,
					Street2: "",
					City:    address.City,
					State:   address.RegionPostalCode,
					Zip:     address.Postcode,
					Country: address.Country,
					Phone:   address.HouseNumber,
					Email:   address.Recipient.Email,
				},
				Parcels: []model.Parcel{
					{
						Length: 5,
						Width:  5,
						Height: 5,
						Weight: 10,
					},
				},
				Items:              items,
				TransactionID:      purchase.Id,
				Date:               purchase.CreatedAt.Unix(),
				Currency:           purchase.Items[0].Currency,
				ContentDescription: fmt.Sprintf("%d items with total value of %f", count, total),
				Status:             purchase.Status.String(),
				APIOrdersID:        int64(s.apiOrderID), //TODO: what is this?
				Incoterm:           "DAP",               //TODO: What is this? Default->DAP
				PaymentMethod:      "Credit Card",       //TODO: this was empty in tests
				ShippingService:    "UPS",               //TODO: What do we use?
				Note:               note,

				Total:              total,
				ItemsCount:         count,
				ShipmentAmountPaid: shipmentFee,
			},
		}
		body, err := json.Marshal(&bodyStruct)
		if err != nil {
			log.Error().Err(err).
				Msg("marshalling the request body")
			continue
		}
		log.Info().RawJSON("raw request body", body).Send()
		req, _ := http.NewRequest("POST", "https://www.shippypro.com/api", bytes.NewBuffer(body))

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", s.integrationKey))

		resp, err := client.Do(req)
		if err != nil {
			log.Error().Err(err).
				Msg("sending request to shippypro")
			continue
		}

		if _, err := ioutil.ReadAll(resp.Body); err != nil {
			log.Err(err).Msg("readingresponse body")
		}
		if resp.StatusCode != http.StatusOK {
			log.Err(fmt.Errorf("non 200 status code")).
				Msgf("got status code %d", resp.StatusCode)
		}
		resp.Body.Close()
	}
}
