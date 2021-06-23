package db

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/veganbase/backend/services/user-service/model"
)

// Info gets minimal information about a list of users (just name
// and email address) and organisations () given their IDs.
func (pg *PGClient) Info(ids []string) (map[string]model.Info, error) {
	retval := map[string]model.Info{}

	userQuery, userArgs, err := sqlx.In(userInfo, ids)
	if err != nil {
		return nil, err
	}
	userQuery = pg.DB.Rebind(userQuery)
	userResults := []struct {
		ID          string  `db:"id"`
		DisplayName *string `db:"display_name"`
		Email       *string  `db:"email"`
		Avatar      *string `db:"avatar"`
	}{}

	err = pg.DB.Select(&userResults, userQuery, userArgs...)
	if err != nil {
		return nil, err
	}

	for _, r := range userResults {
		retval[r.ID] = model.Info{
			ID:    r.ID,
			Name:  r.DisplayName,
			Email: r.Email,
			Image: r.Avatar,
		}
	}

	orgQuery, orgArgs, err := sqlx.In(orgInfo, ids, ids)
	if err != nil {
		return nil, err
	}
	orgQuery = pg.DB.Rebind(orgQuery)
	orgResults := []struct {
		ID    string  `db:"id"`
		Slug  *string `db:"slug"`
		Name  *string `db:"name"`
		Email *string `db:"email"`
		Logo  *string `db:"logo"`
	}{}

	err = pg.DB.Select(&orgResults, orgQuery, orgArgs...)
	if err != nil {
		return nil, err
	}

	for _, r := range orgResults {
		retval[r.ID] = model.Info{
			ID:    r.ID,
			Slug:  r.Slug,
			Name:  r.Name,
			Email: r.Email,
			Image: r.Logo,
		}
	}

	return retval, nil
}

const userInfo = `SELECT id, display_name, email, avatar FROM users WHERE id IN (?)`
const orgInfo = `SELECT id, slug, name, email, logo FROM orgs WHERE id IN (?) OR slug IN (?)`

//GetDeliveryFees is used internally to get delivery fees of multiple users/orgs with one request
func (pg *PGClient) GetDeliveryFees(ids []string) (map[string]model.DeliveryFees, error) {
	returnVal := map[string]model.DeliveryFees{}

	feesQuery, feesArgs, err := sqlx.In(deliveryFeesBy + `owner IN (?)`, ids)
	if err != nil {
		return nil, err
	}
	feesQuery = pg.DB.Rebind(feesQuery)
	fees := []model.DeliveryFees{}

	if err = pg.DB.Select(&fees, feesQuery, feesArgs...); err != nil && err != sql.ErrNoRows{
		return nil, err
	}

	for _, r := range fees {
		returnVal[r.Owner] = r
	}

	return returnVal, nil
}
