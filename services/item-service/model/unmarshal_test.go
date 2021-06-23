package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	category_events "github.com/veganbase/backend/services/category-service/events"
	category_mocks "github.com/veganbase/backend/services/category-service/mocks"
)

var (
	// Just include one category for testing: categories missing from
	// the schema load don't get validated...
	amenities = category_events.Category{
		"car-park":         "car park",
		"reception-24hr":   "24-hour reception",
		"free-wifi":        "free WIFI",
		"restaurant":       "restaurant",
		"breakfast":        "breakfast",
		"lunch":            "lunch",
		"dinner":           "dinner",
		"bar":              "bar",
		"gym":              "gym",
		"pool":             "swimming pool",
		"sport-facilities": "sports facilities",
		"hot-tub":          "hot tub",
		"group-friendly":   "good for groups",
		"events-friendly":  "good for events",
		"pets-allowed":     "pets allowed 🐾",
		"smoking-allowed":  "smoking allowed",
		"full-vegan":       "100% vegan <VB logo>",
	}

	dishCategories = category_events.Category{
		"smoothies": "Smoothies",
		"breakfast": "Breakfast",
	}

	freeFrom = category_events.Category{
		"gluten-free": "Gluten-free",
		"sugar-free":  "Sugar-free",
		"soy-free":    "Soy-free",
	}
	categoryMap = category_events.CategoryMap{
		"venue-amenity": &amenities,
		"free-from":&freeFrom,
		"dish-category": &dishCategories,
	}
)

func TestUnmarshalGood(t *testing.T) {
	c := category_mocks.Client{}
	c.On("Categories").Return(categoryMap)
	for n := range amenities {
		c.On("IsValidLabel", "venue-amenity", n).Return(true)
	}
	for n := range dishCategories {
		c.On("IsValidLabel", "dish-category", n).Return(true)
	}
	for n := range freeFrom {
		c.On("IsValidLabel", "free-from", n).Return(true)
	}
	c.On("IsValidLabel", "venue-amenity", mock.Anything).Return(false)
	c.On("IsValidLabel", "dish-category", mock.Anything).Return(false)
	c.On("IsValidLabel", "free-from", mock.Anything).Return(false)

	LoadSchemas(&c)

	d, _ := os.Open("../testdata/good")
	defer d.Close()
	names, _ := d.Readdirnames(0)

	for _, n := range names {
		b, _ := ioutil.ReadFile("../testdata/good/" + n)
		item := Item{}
		err := json.Unmarshal(b, &item)
		assert.NotEqual(t, item.ItemType, UnknownItem)
		assert.Nil(t, err)
	}
}

func TestUnmarshalBad(t *testing.T) {
	c := category_mocks.Client{}
	c.On("Categories").Return(categoryMap)
	LoadSchemas(&c)

	d, _ := os.Open("../testdata/bad")
	defer d.Close()
	names, _ := d.Readdirnames(0)

	for _, n := range names {
		b, _ := ioutil.ReadFile("../testdata/bad/" + n)
		item := Item{}
		err := json.Unmarshal(b, &item)
		assert.NotNil(t, err)
		fmt.Println(n)
		fmt.Println(err)
		fmt.Println("")
	}
}
