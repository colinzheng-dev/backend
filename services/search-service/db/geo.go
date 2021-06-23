package db
import (
	"errors"
	itemModel "github.com/veganbase/backend/services/item-service/model"
	itemTypes "github.com/veganbase/backend/services/item-service/model/types"
)
// Geo performs a geo-search within a given distance of a latitude,
// longitude point, returning a list of item IDs in order of distance.
func (pg *PGClient) Geo(lat, lon float64, dist float64,
	itemType *itemModel.ItemType,
	approval *[]itemTypes.ApprovalState) ([]string, error) {
	ids := []string{}
	err := pg.DB.Select(&ids,
		geo+typeApprovalWhere(itemType, approval)+geoOrder,
		lon, lat, dist*1000.0)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

const geo = `
SELECT item_id
  FROM item_locations
 WHERE ST_Distance(location, ST_MakePoint($1, $2)) < $3`
const geoOrder = ` ORDER BY ST_Distance(location, ST_MakePoint($1, $2))`


// Region performs a named region search, returning a list of item IDs
// lying within the boundaries of the region.
func (pg *PGClient) Region(name string) ([]string, error) {
	return nil, errors.New("NOT YET IMPLEMENTED")
}

// AddGeo adds geolocation information to the search index for an
// item.
func (pg *PGClient) AddGeo(id string,
	itemType itemModel.ItemType,
	approval itemTypes.ApprovalState,
	latitude, longitude float64) error {
	result, err := pg.DB.Exec(addGeo, id, itemType, approval, longitude, latitude)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("couldn't update location index entry for '" + id + "'")
	}
	return nil
}

const addGeo = `
INSERT INTO item_locations (item_id, item_type, approval, location)
 VALUES ($1, $2, $3, ST_MakePoint($4, $5))
 ON CONFLICT (item_id)
 DO UPDATE SET item_type = $2, approval = $3, location = ST_MakePoint($4, $5)`

