package model

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	it "github.com/veganbase/backend/services/item-service/model"

)

func TestUnmarshalGood(t *testing.T) {

	d, _ := os.Open("../testdata/good")
	defer d.Close()
	names, _ := d.Readdirnames(0)

	for _, n := range names {
		b, _ := ioutil.ReadFile("../testdata/good/" + n)
		item := CartItem{}
		err := json.Unmarshal(b, &item)
		assert.NotEqual(t, item.CartItemFixed.Type, it.UnknownItem)
		assert.Nil(t, err)
	}
}

func TestUnmarshalBad(t *testing.T) {
	d, _ := os.Open("../testdata/bad")
	defer d.Close()
	names, _ := d.Readdirnames(0)

	for _, n := range names {
		b, _ := ioutil.ReadFile("../testdata/bad/" + n)
		item := CartItem{}
		err := json.Unmarshal(b, &item)
		assert.NotNil(t, err)
		fmt.Println(n)
		fmt.Println(err)
		fmt.Println("")
	}
}
