package server

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/veganbase/backend/services/item-service/db"
	"github.com/veganbase/backend/services/item-service/model"
)

// ProcessLinks retrieves link information for an item. The return
// value is an array of map[string]interface{} values that can be
// serialised to JSON.
func (s *Server) ProcessLinks(item *model.ItemFull, linkInfo LinkInfo, userSession string) ([]interface{}, error) {

	// Get all valid link types originating from item.
	originLinkTypes, err := s.db.LinkTypesByOrigin(item.ItemType)
	if err != nil {
		return nil, err
	}
	allowedLinkTypes := map[string]*model.LinkType{}
	for _, linkType := range originLinkTypes {
		allowedLinkTypes[linkType.Name] = &linkType
	}

	// Check requested link types against allowed types and make a set
	// of link types for checking against.
	for name := range linkInfo {
		if _, ok := allowedLinkTypes[name]; !ok {
			log.Error().Msgf("disallowed type %s for item %s, valid types are %v", name, item.ID, allowedLinkTypes)
			return nil, errors.New("disallowed link type '" + name + "' for item")
		}
	}

	// Get all links for origin item.
	links, _, err := s.db.LinksByOriginID(item.ID, 1, 1000)
	if err != nil {
		return nil, err
	}

	// Create lists of links to include in the result, separated by
	// format.
	idLinks := []model.Link{}
	summaryLinks := []model.Link{}
	summaryIDs := []string{}
	fullLinks := []model.Link{}
	fullIDs := []string{}
	for _, link := range *links {
		if format, ok := linkInfo[link.LinkType]; ok {
			switch format {
			case IDsFormat:
				idLinks = append(idLinks, link)
			case SummaryFormat:
				summaryLinks = append(summaryLinks, link)
				summaryIDs = append(summaryIDs, link.Target)
			case FullFormat:
				fullLinks = append(fullLinks, link)
				fullIDs = append(fullIDs, link.Target)
			}
		}
	}

	// Retrieve item information for links that require it.
	params := &db.SearchParams{}
	summaryItems := []*model.ItemSummary{}
	if len(summaryIDs) > 0 {
		summaryItems, _, err = s.db.SummaryItems(params, summaryIDs, []string{}, nil)
		if err != nil {
			return nil, err
		}
	}
	fullItems := []*model.ItemFull{}
	if len(fullIDs) > 0 {
		rawFullItems, _, err := s.db.FullItems(params, fullIDs, []string{}, nil)
		if err != nil {
			return nil, err
		}
		fullItems, err = s.ExpandItemViews(rawFullItems, userSession)
		if err != nil {
			return nil, err
		}
	}

	// Organise linked items for making link views.
	summaries := map[string]*model.ItemSummary{}
	for _, it := range summaryItems {
		summaries[it.ID] = it
	}
	fulls := map[string]*model.ItemFull{}
	for _, it := range fullItems {
		fulls[it.ID] = it
	}

	// Make views for links.
	linkViews := []interface{}{}
	for _, link := range idLinks {
		linkViews = append(linkViews, &link)
	}
	for _, link := range summaryLinks {
		view := model.LinkWithSummary(&link, summaries[link.Target])
		linkViews = append(linkViews, &view)
	}
	for _, link := range fullLinks {
		view := model.LinkWithFull(&link, fulls[link.Target])
		linkViews = append(linkViews, &view)
	}

	return linkViews, nil
}
