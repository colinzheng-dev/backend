package server

import (
	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/events"
)

// Determine allowed owners parameter for authentication context. If
// the user is not an administrator, they can only update items that
// they own or that are owned by organisations of which they are
// members.
func (s *Server) allowedOwners(authInfo *chassis.AuthInfo) ([]string, error) {
	if authInfo.UserIsAdmin {
		return []string{}, nil
	}
	return s.possibleOwners(authInfo.UserID)
}

func (s *Server) possibleOwners(userID string) ([]string, error) {
	if userID == "" {
		return []string{}, nil
	}
	owners := []string{userID}
	orgs, err := s.userSvc.OrgsForUser(userID)
	if err != nil {
		return nil, err
	}
	for orgID := range orgs {
		owners = append(owners, orgID)
	}
	return owners, nil
}

// Emit a pub/sub message notifying of an item change.
func (s *Server) emit(eventType events.ItemEventType, itemID string) {
	msg := events.ItemEvent{
		EventType: eventType,
		ItemID:    itemID,
	}
	chassis.Emit(s, events.ItemChange, msg)
}

func (s *Server) emitWithCollection(eventType events.ItemEventType, itemID, collectionID string) {
	msg := events.ItemEvent{
		EventType:    eventType,
		ItemID:       itemID,
		CollectionID: collectionID,
	}
	chassis.Emit(s, events.ItemChange, msg)
}

func (s *Server) getOrgId(idOrSlug string) (*string, error) {
	//TODO: THIS CHECKING IS A BIT ALL OVER THE PLACE, REFACTOR IT LATER
	orgs, err := s.userSvc.Info([]string{idOrSlug})
	if err != nil {
		return nil, err
	}
	var orgId string
	for _, v := range orgs {
		orgId = v.ID
	}
	return &orgId, nil
}
