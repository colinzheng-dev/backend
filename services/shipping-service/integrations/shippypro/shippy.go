package shippypro

import (
	"github.com/veganbase/backend/chassis"
)

type Shippo struct {
	BaseURL        string
	IntegrationKey string
	IsDevMode      bool
}

func New(baseUrl, integrationKey string, isDev bool) (*Shippo, error) {
	_ = chassis.CheckURL(baseUrl, "shippypro url")

	integration := Shippo{
		BaseURL:        baseUrl,
		IsDevMode:      isDev,
		IntegrationKey: integrationKey,
	}

	return &integration, nil
}
