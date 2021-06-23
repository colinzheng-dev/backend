package mailer

import (
	"errors"
	"regexp"
	"strconv"
	"sync"
	"time"

	mailjet "github.com/mailjet/mailjet-apiv3-go"
	"github.com/mailjet/mailjet-apiv3-go/resources"
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/services/email-service/model"
	site_model "github.com/veganbase/backend/services/site-service/model"
)

var fromRE = regexp.MustCompile(`(.+) <([^>]+)>`)

// MailjetMailer sends email using Mailjet.
type MailjetMailer struct {
	mj    *mailjet.Client
	mu    sync.Mutex
	tmpls map[string]templateInfo
}

// Fields we need to extract from the template definition to use the
// sending API.
type templateInfo struct {
	id          int64
	fromEmail   string
	fromName    string
	replyTo     string
	replyEmail  string
	senderEmail string
	senderName  string
	subject     string
}

// NewMailjetMailer creates a new mailer for Mailjet based on Mailjet
// API key credentials.
func NewMailjetMailer(pubkey, privkey string) (*MailjetMailer, error) {
	mailer := MailjetMailer{}
	log.Info().Msg("connecting to Mailjet")
	mailer.mj = mailjet.NewMailjetClient(pubkey, privkey)
	err := mailer.loadTemplates()
	if err != nil {
		return nil, err
	}
	go mailer.templateUpdater()
	return &mailer, nil
}

// loadTemplates loads all our template from Mailjet.
func (m *MailjetMailer) loadTemplates() error {
	// Get basic template information first.
	var tmpls []resources.Template
	_, _, err := m.mj.List("template", &tmpls, mailjet.Filter("OwnerType", "user"))
	if err != nil {
		log.Error().Err(err).Msg("failed to load templates from Mailjet")
		return err
	}

	// Now look up the detailed information for each template (have to
	// do these one at a time).
	tmplInfo := map[string]templateInfo{}
	for _, tmpl := range tmpls {
		req := mailjet.Request{
			Resource: "template",
			ID:       tmpl.ID,
			Action:   "detailcontent",
		}

		// Get detailed template information and make sure it matches what
		// we expect to see.
		var res []resources.TemplateDetailcontent
		err = m.mj.Get(&req, &res)
		if err != nil {
			log.Error().Err(err).
				Str("name", tmpl.Name).
				Msg("getting template details")
			continue
		}
		if len(res) != 1 {
			log.Error().
				Str("name", tmpl.Name).
				Msg("strange reply from Mailjet for template details")
			continue
		}
		headerMap, ok := res[0].Headers.(map[string]interface{})
		if !ok {
			log.Error().
				Str("name", tmpl.Name).
				Msg("invalid headers in template details from Mailjet")
			continue
		}

		// The "From" header inconveniently includes the From email and
		// name together. We need them separately for sending mail.
		from := fromRE.FindStringSubmatch(headerMap["From"].(string))

		// Collect template information to save for later use when sending
		// mail.
		tmplInfo[tmpl.Name] = templateInfo{
			id:          tmpl.ID,
			fromEmail:   from[2],
			fromName:    from[1],
			replyTo:     headerMap["Reply-To"].(string),
			replyEmail:  headerMap["ReplyEmail"].(string),
			senderEmail: headerMap["SenderEmail"].(string),
			senderName:  headerMap["SenderName"].(string),
			subject:     headerMap["Subject"].(string),
		}
	}

	// Save the template information.
	m.mu.Lock()
	defer m.mu.Unlock()
	oldtmpls := m.tmpls
	m.tmpls = tmplInfo
	for name, tmpl := range m.tmpls {
		old, chk := oldtmpls[name]
		if !chk {
			log.Info().
				Str("name", name).
				Int64("id", tmpl.id).
				Msg("loaded new Mailjet template")
			continue
		}
		if old != tmpl {
			log.Info().
				Str("name", name).
				Int64("id", tmpl.id).
				Msg("updated existing Mailjet template")
		}
	}
	return nil
}

// Repeatedly reload templates.
func (m *MailjetMailer) templateUpdater() {
	for range time.Tick(3 * time.Minute) {
		m.loadTemplates()
	}
}

// Send sends an email using a Mailjet template.
func (m *MailjetMailer) Send(topic *model.TopicInfo, site *site_model.Site,
	language string, data map[string]interface{}) error {
	// Get Mailjet template.
	var tmplInfo *templateInfo
	ok := false
	if site != nil {
		tmplInfo, ok = m.getTemplate(topic.Name + ":" + site.ID)
	}
	if !ok {
		tmplInfo, ok = m.getTemplate(topic.Name)
		if !ok {
			return ErrUnknownEmailTemplate
		}
	}

	// Add site information to email variables.
	var emailDomain string
	if site != nil {
		data["site_name"] = site.Name
		data["site_url"] = site.URL
		data["site_signature"] = site.Signature
		emailDomain = site.EmailDomain
	} else {
		data["site_name"] = "a Veganlogin site"
		data["site_url"] = "https://veganlogin.com"
		data["site_signature"] = "Veganlogin"
		emailDomain = "veganlogin.com"
	}

	// Build mail information and send mail.
	sendAddress := topic.SendAddress
	if sendAddress == "" {
		sendAddress = "info"
	}
	info := &mailjet.InfoSendMail{
		FromEmail: sendAddress + "@" + emailDomain,
		FromName:  tmplInfo.fromName,
		Recipients: []mailjet.Recipient{
			{
				Email: data["email"].(string),
			},
		},
		Subject:            tmplInfo.subject,
		MjTemplateLanguage: "true",
		MjTemplateID:       strconv.Itoa(int(tmplInfo.id)),
		MjTemplateErrorReporting: "rob.fuinha@gmail.com",
		MjTemplateErrorDeliver: "true",
		Vars:               data,
	}
	res, err := m.mj.SendMail(info)
	if err != nil {
		return err
	}
	if len(res.Sent) != 1 {
		err = errors.New("invalid result from Mailjet")
		log.Error().Err(err)
		return err
	}

	// Log mail send.
	// TODO: RECORD EVENT HERE!
	log.Info().
		Int64("mailjet-message-id", res.Sent[0].MessageID).
		Str("topic", topic.Name).
		Msg("email sent")
	return nil
}

// Look up a template ID from its name in our map.
func (m *MailjetMailer) getTemplate(name string) (*templateInfo, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tmpls[name]
	return &t, ok
}
