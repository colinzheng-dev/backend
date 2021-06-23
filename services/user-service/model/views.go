package model

// UserPublicProfile is a restriction of the User model used to return
// public profile information.
type UserPublicProfile struct {
	ID          string              `json:"id"`
	DisplayName *string             `json:"display_name"`
	Avatar      *string             `json:"avatar,omitempty"`
	Country     *string             `json:"country,omitempty"`
	Orgs        *[]*OrgWithUserInfo `json:"orgs,omitempty"`
}

// UserFullProfile is the full information about a user plus the
// organisations the user belongs to.
type UserFullProfile struct {
	User
	Orgs *[]*OrgWithUserInfo `json:"orgs,omitempty"`
}

// Info is a minimal view of user or organisation information used in
// various places as a helper.
type Info struct {
	ID    string  `json:"id"`
	Slug  *string `json:"slug,omitempty"`
	Name  *string `json:"name"`
	Email *string `json:"email"`
	Image *string `json:"image"`
}

// OrgWithUserInfo is a view of an organisation including an extra
// flag to say whether the associated user is an organisation
// administrator.
type OrgWithUserInfo struct {
	Organisation
	UserIsAdmin bool `db:"is_org_admin" json:"user_is_admin"`
}

// UserWithOrgInfo is a view of a user that includes a flag to mark
// whether the user is an organisation administrator for an
// organisation.
type UserWithOrgInfo struct {
	Info
	UserIsAdmin bool `json:"user_is_admin"`
}

// UserWithOrgAdminFlag creates a user information view including an
// organisation administrator flag.
func UserWithOrgAdminFlag(user *User, isAdmin bool) *UserWithOrgInfo {
	result := UserWithOrgInfo{}
	result.ID = user.ID
	result.Email = &user.Email
	if user.Name != nil {
		result.Name = user.Name
	}
	if user.Avatar != nil {
		result.Image = user.Avatar
	}
	result.UserIsAdmin = isAdmin
	return &result
}

// InfoFromOrg creates a brief view of an organisation.
func InfoFromOrg(org *Organisation) *Info {
	return &Info{
		ID:    org.ID,
		Slug:  &org.Slug,
		Name:  &org.Name,
		Email: &org.Email,
		Image: &org.Logo,
	}
}


