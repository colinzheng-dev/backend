package messages

// OrgAddUser is the request body used to add users to an
// organisation.
type OrgAddUser struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	IsOrgAdmin bool   `json:"is_org_admin"`
}

// OrgPatchUser is the request body used to modify the organisation
// admin status of users in an organisation.
type OrgPatchUser struct {
	IsOrgAdmin bool `json:"is_org_admin"`
}
