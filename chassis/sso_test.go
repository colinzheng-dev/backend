package chassis

import (
	"encoding/json"
	"testing"
	"time"
)

const secret = "4c1ab18fc0e584533fce53339791801d"

// The customer information is represented as a hash which must contain at least the email address
// of the customer and a current timestamp (in ISO8601 encoding).
// You can also include the customer's first name, last name or several shipping addresses.
// Optionally, you can include an IP address of the customer's current browser session,
// that makes the token valid only for requests originating from this IP address.
// You can attribute tags to your customer by setting "tag_string" to a list of comma separated one-word values.
// These tags will override any tags that you may have already attributed to this customer.
// At Shopify, we use email addresses as unique identifiers for customers of a shop.

// If your site uses other identifiers (such as usernames),
// or if it is possible that two different users of your site registered with the same email address,
// you must set the unique identifier in the "identifier" field to avoid security problems.

// If the email address is always unique, you don't need to set the "identifier" field.

// If you want your users to see a specific page of your Shopify store, you can use the "return_to" field for that.
type SSORequest struct {
	Email     string     `json:"email"`
	CreatedAt string     `json:"created_at"`
	Attrs     GenericMap `json:"attrs"`
}

func TestTokenGenerate(t *testing.T) {
	//TODO: THIS IS ONLY TESTING THE WHOLE PROCESS
	request := &SSORequest{
		Email:     "bob@shopify.com",
		CreatedAt: time.Now().Format("2006-01-02T15:04:05-07:00"),
	}
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		t.Errorf("failed to marshal sso request: %s", err.Error())
		return
	}

	token, err := GenerateToken(secret, jsonBytes)
	if err != nil {
		t.Errorf("token generate failed: %s", err.Error())
		return
	}

	revertedBytes, err := RevertToken(secret, *token)
	if err != nil {
		t.Errorf("token reversion failed: %s", err.Error())
		return
	}
	var revertedRequest SSORequest
	if err = json.Unmarshal(*revertedBytes, &revertedRequest); err != nil {
		t.Errorf("error unmarshalling sso request: %s", err.Error())
		return
	}
	if request.Email != revertedRequest.Email || request.CreatedAt != revertedRequest.CreatedAt {
		t.Error("reverted request is different from original")
		return
	}

}
