package server

import (
	"errors"
)

//checks if the user is admin of an org
func (s *Server) userIsOrgAdmin(userId, orgId string) error {
	var isOrgAdmin bool
	//getting all users of the org
	orgUsers, err := s.db.OrgUsers(orgId)
	if err != nil {
		return err
	}
	//check if one matches the given user
	for _, u := range orgUsers {
		if u.UserID == userId {
			isOrgAdmin = u.IsOrgAdmin
			break
		}
	}

	if isOrgAdmin {
		return nil
	}
	return errors.New("user is not admin of the owning org")
}