package mailer

import (
	"fmt"

	"github.com/veganbase/backend/services/email-service/model"
	site_model "github.com/veganbase/backend/services/site-service/model"
)

type DevMailer struct{}

func NewDevMailer() *DevMailer {
	return &DevMailer{}
}

func (m *DevMailer) Send(topic *model.TopicInfo, site *site_model.Site,
	language string, data map[string]interface{}) error {
	fmt.Println("====> EMAIL SEND")
	fmt.Println("  topic =", topic.Name, "   site =", site, "   language =", language)
	for k, v := range data {
		fmt.Println(" ", k, "=", v)
	}
	fmt.Println("<==== EMAIL SEND")
	return nil
}
