package server

import (
	"github.com/veganbase/backend/services/item-service/model"
)

// ExpandItemViews makes full views from raw items from the database,
// supplementing the views with user information.
func (s *Server) ExpandItemViews(rawItems []*model.ItemWithStatistics, userSession string) ([]*model.ItemFull, error) {
	// Extract unique user IDs involved in items and get information
	// about those users. Also get a slice with all item_ids to be passed to social service.
	userIDs := map[string]bool{}
	var itemIDs []string
	for _, it := range rawItems {
		userIDs[it.Creator] = true
		userIDs[it.Owner] = true
		itemIDs = append(itemIDs, it.ID)
	}

	uniqueUserIDs := []string{}
	for k := range userIDs {
		uniqueUserIDs = append(uniqueUserIDs, k)
	}
	userInfo, err := s.userSvc.Info(uniqueUserIDs)
	if err != nil {
		return nil, err
	}

	var upvotes *map[string]bool
	if userSession != "" {
		if upvotes, err = s.socialSvc.GetUserUpvotes(userSession); 	err != nil {
			return nil, err
		}
	}

	collNames, err := s.db.CollectionsNamesByItemId(itemIDs)

	if err != nil {
		return nil, err
	}
	// Make item views.
	items := []*model.ItemFull{}
	for _, it := range rawItems {
		var userUpvoted bool
		if upvotes != nil {
			userUpvoted = (*upvotes)[it.ID]
		}
		items = append(items, model.FullView(&it.Item, userInfo[it.Creator], userInfo[it.Owner], it.Upvotes, userUpvoted, it.Rank, (*collNames)[it.ID]))
	}

	return items, nil
}
