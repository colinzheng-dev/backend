package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/db"
	"github.com/veganbase/backend/services/item-service/model/types"
	"github.com/veganbase/backend/services/item-service/server"
	content "google.golang.org/api/content/v2.1"
	"google.golang.org/api/option"
)

var dbURL = flag.String("db-url", "", "item-service database url")
var live = flag.Bool("live", false, "push data into the API. false will just print the items")
var only = flag.String("only", "", "only update this product")

func main() {
	flag.Parse()

	ctx := context.Background()

	dbClient, err := db.NewPGClient(ctx, *dbURL)
	if err != nil {
		fmt.Println(err)
		panic("couldn't connect to database")
	}

	merchantID := uint64(417663490)

	svc, err := content.NewService(ctx, option.WithCredentialsFile("/tmp/content-api-key.json"))
	if err != nil {
		panic(err)
	}

	collection, err := dbClient.CollectionViewByName("_mightyplants_shopping")
	if err != nil {
		panic(err)
	}

	apiClient := content.NewProductsService(svc)

	pageSize := uint(100)

	search := &db.SearchParams{
		Approval: &[]types.ApprovalState{types.Approved},
	}

	var total = uint(1)
	for i := uint(0); i < total; i += pageSize {
		items, t, err := dbClient.FullItems(search, nil, collection.IDs, &chassis.Pagination{Page: i/pageSize + 1, PerPage: pageSize})
		if err != nil {
			panic(err)
		}

		if t != nil {
			total = *t
		}

		for _, item := range items {

			if *only != "" && item.ID != *only {
				continue
			}

			if *live {
				product, err := server.AddProductToContentAPI(dbClient, apiClient, merchantID, &item.Item)
				if err != nil {
					fmt.Println("failed", item.ID, item.CreatedAt, err)
					continue
				}
				fmt.Println(item.ID, item.CreatedAt, "product", product.Id)
			} else {
				fmt.Println(item.ID, item.CreatedAt, item.Attrs)
			}
		}
	}
}
