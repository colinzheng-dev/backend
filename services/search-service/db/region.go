package db

import (
	"database/sql"
	"fmt"
	itemModel "github.com/veganbase/backend/services/item-service/model"
	itemTypes "github.com/veganbase/backend/services/item-service/model/types"
	"github.com/veganbase/backend/services/search-service/model"
	"strconv"
	"strings"
)

func (pg *PGClient) Countries() (*[]model.Region, error) {
	countries := []model.Region{}
	if err := pg.DB.Select(&countries, qGetRegions + ` region_type = 'country' AND iso_a2 IS NOT NULL ORDER BY name ASC`); err != nil {
		return nil, err
	}
	return &countries, nil
}

func (pg *PGClient) CountryByID(countryID string) (*model.Region, error) {
	country := model.Region{}
	if err := pg.DB.Get(&country, qGetRegions + ` region_type = 'country'` + `AND iso_a2 = '` + countryID + `'` ); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCountryNotFound
		}
		return nil, err
	}
	return &country, nil
}

func (pg *PGClient) States(countryID string) (*[]model.Region, error) {
	states := []model.Region{}
	q := qGetRegions + ` region_type = 'state' AND iso_a2 = '` + countryID + `' ORDER BY name ASC `

	if err := pg.DB.Select(&states, q); err != nil {
		return nil, err
	}
	return &states, nil
}

func (pg *PGClient) StateByID(stateID string) (*model.Region, error) {
	state := model.Region{}
	if err := pg.DB.Get(&state, qGetRegions + ` region_type = 'state'` + ` AND iso_a2 = '` + stateID + `'` ); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrStateNotFound
		}
		return nil, err
	}
	return &state, nil
}

const qGetRegions = `
	SELECT id, region_type, iso_a2, iso_3166_2, postal, name, adm0_a3, latitude, longitude
	FROM default_regions
	WHERE 
`

func (pg *PGClient) IsInsideRegions(latitude, longitude float64, references []int ) (*bool, error){
	var result bool
	// we build this query this way because ST_Intersects and ST_GeomFromText
	// are tricky to work with query arguments
	q := injectCoordinates(qIsInsideRegion, latitude, longitude) + buildInClause(references)
	if err := pg.DB.Get(&result, q); err != nil && err != sql.ErrNoRows{
		return nil, err
	}
	return &result, nil
}

const qIsInsideRegion = `
	SELECT 1 FROM default_regions dr
	WHERE ST_Intersects(dr.geom::geography, ST_GeomFromText('POINT(<lng> <lat>)',4326)) AND dr.id IN `

func injectCoordinates(q string, lat, long float64) string {
	q = strings.ReplaceAll(q, "<lng>", fmt.Sprintf("%f", long))
	return strings.ReplaceAll(q, "<lat>", fmt.Sprintf("%f", lat))
}

func buildInClause(arr []int) string {
	asString := []string{}
	for _, v :=range arr {
		asString = append(asString, strconv.Itoa(v))
	}
	return "(" + strings.Join(asString, ",") + ")"
}

func (pg *PGClient) GetItemsInsideRegion(regionType, regionReference string, itemType *itemModel.ItemType,
	approval *[]itemTypes.ApprovalState ) ([]string, error) {
	ids := []string{}
	q := qItemsByRegion + typeApprovalWhere(itemType, approval) + " AND " +  getWhereByRegionType(regionType, regionReference)

	if err := pg.DB.Select(&ids, q); err != nil {
		return nil, err
	}
	return ids, nil
}

const qItemsByRegion = `
	SELECT item_id
 	FROM item_locations, default_regions 
	WHERE ST_Intersects(geom, location) 
`

func getWhereByRegionType(regionType, regionReference string) string {
	switch regionType {
	case "country":
		return "iso_a2 = '"+ regionReference + "' AND region_type = 'country'"
	case "state":
		return "iso_3166_2 = '"+ regionReference + "' AND region_type = 'state'"
	}
	return ""
}
