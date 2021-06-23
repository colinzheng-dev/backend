package server

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/chassis/pubsub"
	"github.com/veganbase/backend/services/item-service/db"
	"github.com/veganbase/backend/services/item-service/events"
	"github.com/veganbase/backend/services/item-service/model"
	content "google.golang.org/api/content/v2.1"
	"google.golang.org/api/option"
)

func (s *Server) contentAPIClient() (*content.ProductsService, uint64, error) {
	var opts []option.ClientOption
	if s.contentAPICredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(s.contentAPICredentialsFile))
	}

	svc, err := content.NewService(context.Background(), opts...)
	if err != nil {
		return nil, 0, err
	}

	merchantID := uint64(417663490)

	return content.NewProductsService(svc), merchantID, nil
}

func (s *Server) HandleItemCreateOrUpdate() {

	client, merchantID, err := s.contentAPIClient()
	if err != nil {
		log.Error().Err(err).Msg("unable to configure content API in HandleItemCreateOrUpdate")
		return
	}

	ch, _, err := s.PubSub.Subscribe(string(events.ItemChange), s.AppName, pubsub.Fanout)
	if err != nil {
		log.Fatal().Err(err).
			Msg("unable to subscribe to user update topic")
	}

	for {
		raw := <-ch

		evt := events.ItemEvent{}
		if err := json.Unmarshal(raw, &evt); err != nil {
			log.Error().Err(err).
				Msg("decoding item create message")
			continue
		}

		if evt.EventType != events.ItemCreated && evt.EventType != events.ItemUpdated && evt.EventType != events.ItemAddedToCollection {
			continue
		}

		item, err := s.db.ItemByID(evt.ItemID)
		if err != nil {
			log.Error().Err(err).
				Msgf("retrieving item %s to update in google content API", evt.ItemID)
			continue
		}

		// XXX: hardcoded to only look for items in `_mightyplants_shopping`
		// if we want to expand this will need to break out this service to load
		// a set of listeners and their associated merchantIDs
		collection, err := s.db.CollectionViewByName("_mightyplants_shopping")
		if err != nil {
			log.Error().Err(err).
				Msgf("loading collection %s", evt.ItemID)
			continue
		}

		sort.Strings(collection.IDs)
		if sort.SearchStrings(collection.IDs, item.ID) == len(collection.IDs) {
			log.Info().
				Msgf("item is not in the collection %s", evt.ItemID)
			continue
		}

		product, err := AddProductToContentAPI(s.db, client, merchantID, item)
		if err != nil {
			log.Error().Err(err).
				Msgf("error creating or updating product %s in google content API: %s", item.ID, err)
			continue
		}

		log.Info().Msgf("synced item %s to google content API, product: %s", item.ID, product.Id)
	}
}

func (s *Server) HandleItemDeletion() {

	client, merchantID, err := s.contentAPIClient()
	if err != nil {
		log.Error().Err(err).Msg("unable to configure content API in HandleItemDeletion")
		return
	}

	ch, _, err := s.PubSub.Subscribe(string(events.ItemChange), s.AppName, pubsub.Fanout)
	if err != nil {
		log.Fatal().Err(err).
			Msg("unable to subscribe to user delete topic")
	}

	for {

		raw := <-ch

		evt := events.ItemEvent{}
		if err := json.Unmarshal(raw, &evt); err != nil {
			log.Error().Err(err).
				Msg("decoding item create message")
			continue
		}

		if evt.EventType != events.ItemDeleted && evt.EventType != events.ItemRemovedFromCollection {
			continue
		}

		if evt.CollectionID == "" {
			log.Error().Err(err).
				Msg("can only delete on collection changes")
			continue
		}

		item, err := s.db.ItemByID(evt.ItemID)
		if err != nil {
			log.Error().Err(err).
				Msgf("couldn't load item %s", evt.ItemID)
			continue
		}

		// XXX: hardcoded to only look for items in `_mightyplants_shopping`
		// if we want to expand this will need to break out this service to load
		// a set of listeners and their associated merchantIDs
		collection, err := s.db.CollectionViewByName("_mightyplants_shopping")
		if err != nil {
			log.Error().Err(err).
				Msgf("loading collection %s", evt.ItemID)
			continue
		}

		if strconv.Itoa(collection.ID) != evt.CollectionID {
			log.Info().
				Msgf("not a tracked collection (%s removed from %s)", evt.ItemID, evt.CollectionID)
			continue
		}

		pof, err := getProductOffering(s.db, item)
		if err != nil {
			log.Error().Err(err).
				Msgf("loading product offering for %s", evt.ItemID)
			continue
		}

		sku, ok := pof.Attrs["unique_identifier"].(string)
		if !ok {
			log.Error().Err(err).
				Msgf("unique_identifier in wrong format: %#v", pof.Attrs["unique_identifier"])
			continue
		}

		// XXX: store this on the item instead of harcoding it
		productID := fmt.Sprintf("online:en:GB:%s", sku)
		req := client.Delete(merchantID, productID)

		if err := req.Do(); err != nil {
			log.Error().Err(err).
				Msgf("error deleting product %s (item %s) in google content API", productID, evt.ItemID)
			continue
		}

		log.Info().Msgf("deleted item %s (product %s) from google content API", evt.ItemID, productID)
	}
}

func getProductOffering(dbClient db.DB, product *model.Item) (*model.Item, error) {
	links, _, err := dbClient.LinksByOriginID(product.ID, 1, 25)
	if err != nil {
		return nil, err
	}

	if links == nil {
		return nil, fmt.Errorf("no links found for %s", product.ID)
	}

	var pof *model.Item
	for _, link := range *links {
		// XXX: not sure why this wasn't const'd somewhere
		if link.LinkType == "product-has-offerings" {
			pof, err = dbClient.ItemByID(link.Target)
			if err != nil {
				return nil, err
			}
			break
		}
	}

	if pof == nil {
		return nil, fmt.Errorf("no offering found for product: %s", product.ID)
	}

	return pof, nil
}

func AddProductToContentAPI(dbClient db.DB, client *content.ProductsService, merchantID uint64, product *model.Item) (*content.Product, error) {

	pof, err := getProductOffering(dbClient, product)
	if err != nil {
		return nil, err
	}

	var allImgs = make(map[string]bool)
	for _, pic := range product.Pictures {
		allImgs[pic] = true
	}

	var firstImg string
	if product.FeaturedPicture != "" {
		firstImg = product.FeaturedPicture
		delete(allImgs, firstImg)
	}

	var otherImgs = make([]string, 0, len(allImgs))
	for img := range allImgs {
		otherImgs = append(otherImgs, img)
	}

	price, ok := pof.Attrs["price"].(float64)
	if !ok {
		return nil, fmt.Errorf("price in wrong format: %#v - %v", pof.Attrs["price"], reflect.TypeOf(pof.Attrs["price"]))
	}

	currency, ok := pof.Attrs["currency"].(string)
	if !ok {
		return nil, fmt.Errorf("currency in wrong format: %#v", pof.Attrs["currency"])
	}

	sku, ok := pof.Attrs["unique_identifier"].(string)
	if !ok {
		return nil, fmt.Errorf("unique_identifier in wrong format: %#v", pof.Attrs["unique_identifier"])
	}

	qty, ok := pof.Attrs["available_quantity"].(float64)
	if !ok {
		return nil, fmt.Errorf("available_quantity in wrong format: %#v", pof.Attrs["available_quantity"])
	}

	availability := "in stock"
	if qty <= 0 {
		availability = "out of stock"
	}

	req := client.Insert(merchantID, &content.Product{
		OfferId:              sku,
		Title:                pof.Name,
		Description:          product.Description,
		ContentLanguage:      pof.Lang,
		Channel:              "online",
		Link:                 "https://mightyplants.com/" + product.Slug, // XXX: hardcoded for MightyPlants
		ImageLink:            firstImg,
		AdditionalImageLinks: otherImgs,
		TargetCountry:        "GB", // XXX: hardcoded for MightyPlants
		Price: &content.Price{
			Currency: currency,
			Value:    fmt.Sprintf("%.2f", price/100),
		},
		Availability: availability,
	})

	return req.Do()
}
