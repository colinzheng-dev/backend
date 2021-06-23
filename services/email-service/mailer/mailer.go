package mailer

import (
	"errors"

	"github.com/veganbase/backend/services/email-service/model"
	site_model "github.com/veganbase/backend/services/site-service/model"
)

// ErrUnknownEmailTemplate is the error returned by a template store
// when an unknown template is requested.
var ErrUnknownEmailTemplate = errors.New("email template unknown")

// Mailer represents machinery for sending template-based emails.
type Mailer interface {
	Send(topic *model.TopicInfo, site *site_model.Site,
		language string, data map[string]interface{}) error
}
