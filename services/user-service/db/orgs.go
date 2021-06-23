package db

import (
	"database/sql"
	"github.com/veganbase/backend/services/user-service/messages"
	"strings"

	"github.com/gosimple/slug"

	// Import Postgres DB driver.
	_ "github.com/lib/pq"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/model"
)

// CreateOrg creates a new organisation.
func (pg *PGClient) CreateOrg(org *model.Organisation) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Generate new organisation ID and initial attempt at a slug (which
	// we might have to change to make it unqiue).
	org.ID = chassis.NewID("org")
	org.Slug = slug.Make(org.Name)
	if len(org.Slug) < 8 {
		org.Slug += "-organisation"
	}

	// Repeatedly try to insert the organisation, dealing with potential
	// slug collisions by adding a random string to the slug if the
	// insert fails.
	for {
		rows, err := tx.NamedQuery(qCreateOrg, org)
		if err != nil {
			return err
		}
		if rows.Next() {
			err = rows.Scan(&org.CreatedAt)
			if err != nil {
				return err
			}
			break
		}

		// Slug collision: try again...
		org.Slug = slug.Make(org.Name + " " + chassis.NewBareID(4))
	}

	return err
}

const qCreateOrg = `
INSERT INTO
  orgs
   (id, slug, name, description, logo,
    address, phone, email, urls, industry,
    year_founded, employees)
 VALUES
   (:id, :slug, :name, :description, :logo,
    :address, :phone, :email, :urls, :industry,
    :year_founded, :employees)
 ON CONFLICT DO NOTHING
 RETURNING created_at`

// OrgByID looks up an organisation by its organisation ID.
func (pg *PGClient) OrgByID(id string) (*model.Organisation, error) {
	org := &model.Organisation{}
	err := pg.DB.Get(org, orgByID, id)
	if err == sql.ErrNoRows {
		return nil, ErrOrgNotFound
	}
	if err != nil {
		return nil, err
	}
	return org, nil
}

// OrgByIDorSlug looks up an organisation by its organisation ID or slug.
func (pg *PGClient) OrgByIDorSlug(idOrSlug string) (*model.Organisation, error) {
	org := &model.Organisation{}
	if err := pg.DB.Get(org, orgByID + " OR slug = $1 ", idOrSlug); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return org, nil
}

const orgByID = `
SELECT id, name, slug, logo, description, address, phone, email,
  urls, industry, year_founded, employees, created_at
  FROM orgs
 WHERE id = $1`

// Orgs gets a list of organisations in alphabetical name order,
// optionally filtered by a search term and paginated.
func (pg *PGClient) Orgs(search string, page, perPage uint) ([]*model.Organisation, error) {
	var err error
	results := []*model.Organisation{}
	if search != "" {
		err = pg.DB.Select(&results,
			orgListWithSearch+chassis.Paginate(page, perPage), "%"+strings.ToLower(search)+"%")
	} else {
		err = pg.DB.Select(&results,
			orgList+chassis.Paginate(page, perPage))
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}

const orgList = `
SELECT id, name, slug, logo, description, address, phone, email,
  urls, industry, year_founded, employees, created_at
  FROM orgs
 ORDER BY name`

const orgListWithSearch = `
SELECT id, name, slug, logo, description, address, phone, email,
  urls, industry, year_founded, employees, created_at
  FROM orgs
 WHERE lower(email) LIKE $1 OR lower(name) LIKE $1
 ORDER BY name`

// OrgUsers gets a list of all users in an organisation.
func (pg *PGClient) OrgUsers(id string) ([]*model.OrgUser, error) {
	users := []*model.OrgUser{}
	if err := pg.DB.Select(&users, orgUsers, id); err != nil {
		return nil, err
	}
	return users, nil
}

const orgUsers = `
SELECT id, org_id, user_id, is_org_admin, created_at
  FROM org_users WHERE org_id = $1`

// UpdateOrg updates an organisation in the database.
func (pg *PGClient) UpdateOrg(org *model.Organisation) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	check := &model.Organisation{}
	err = tx.Get(check, orgByID, org.ID)
	if err == sql.ErrNoRows {
		return ErrOrgNotFound
	}
	if err != nil {
		return err
	}

	// Check read-only fields.
	if org.Slug != check.Slug ||
		org.CreatedAt != check.CreatedAt {
		return ErrReadOnlyField
	}

	// Do the update.
	_, err = tx.NamedExec(updateOrg, org)
	if err != nil {
		return err
	}
	return nil
}

const updateOrg = `
UPDATE orgs SET name=:name, description=:description, logo=:logo,
                address=:address, phone=:phone, email=:email,
                urls=:urls, industry=:industry,
                year_founded=:year_founded, employees=:employees
 WHERE id = :id`

// DeleteOrg deletes the given organisation.
// TODO: ADD SOME SORT OF ARCHIVAL MECHANISM INSTEAD.
func (pg *PGClient) DeleteOrg(id string) error {
	result, err := pg.DB.Exec(deleteOrg, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrOrgNotFound
	}
	return nil
}

const deleteOrg = "DELETE FROM orgs WHERE id = $1"

// UserOrgs gets a list of organisations that a user is a member of, in
// alphabetical name order.
func (pg *PGClient) UserOrgs(userID string) ([]*model.OrgWithUserInfo, error) {
	var err error
	results := []*model.OrgWithUserInfo{}
	err = pg.DB.Select(&results, userOrgs, userID)
	if err != nil {
		return nil, err
	}
	return results, nil
}

const userOrgs = `
SELECT o.id, o.name, o.slug, o.logo, o.description, o.address, o.phone, o.email,
  o.urls, o.industry, o.year_founded, o.employees, o.created_at, u.is_org_admin
  FROM orgs o
  JOIN org_users u ON o.id = u.org_id
 WHERE u.user_id = $1
 ORDER BY name`

// OrgAddUser adds a user to an organisation.
func (pg *PGClient) OrgAddUser(orgID string, userID string, isOrgAdmin bool) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	// Check that the organisation and user exist.
	var check int
	err = tx.Get(&check, `SELECT COUNT(id) FROM orgs WHERE id = $1`, orgID)
	if err == sql.ErrNoRows || check != 1 {
		return ErrOrgNotFound
	}
	if err != nil {
		return err
	}
	check = 0
	err = tx.Get(&check, `SELECT COUNT(id) FROM users WHERE id = $1`, userID)
	if err == sql.ErrNoRows || check != 1 {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}

	// Do the insert. Here, a failure to insert indicates that the user
	// is already a member of the organisation.
	res, err := tx.Exec(orgAddUser, orgID, userID, isOrgAdmin)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrUserAlreadyInOrg
	}

	return nil
}

const orgAddUser = `
INSERT INTO org_users (org_id, user_id, is_org_admin)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING`

// OrgDeleteUser removes a user from an organisation.
func (pg *PGClient) OrgDeleteUser(orgID string, userID string) error {
	res, err := pg.DB.Exec(orgDeleteUser, orgID, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrUserNotFound
	}
	return nil
}

const orgDeleteUser = `
DELETE FROM org_users WHERE org_id = $1 AND user_id = $2`

// OrgPatchUser updates a user's admin status within an organisation.
func (pg *PGClient) OrgPatchUser(orgID string, userID string, isOrgAdmin bool) error {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	// Check that the organisation and user exist.
	var check int
	err = tx.Get(&check, `SELECT COUNT(id) FROM orgs WHERE id = $1`, orgID)
	if err == sql.ErrNoRows || check != 1 {
		return ErrOrgNotFound
	}
	if err != nil {
		return err
	}
	check = 0
	err = tx.Get(&check, `SELECT COUNT(id) FROM users WHERE id = $1`, userID)
	if err == sql.ErrNoRows || check != 1 {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}

	// Do the insert. Here, a failure to insert indicates that the user
	res, err := tx.Exec(orgPatchUser, orgID, userID, isOrgAdmin)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrUserNotFound
	}
	return nil
}

const orgPatchUser = `
UPDATE org_users SET is_org_admin = $3 WHERE org_id = $1 AND user_id = $2`


func (pg *PGClient) NotificationInfoByOrgId(orgId string) (*model.EmailNotificationInfo, error) {

	temp := struct {
		DisplayName string `db:"name"`
		Email       string  `db:"email"`
	}{}
	if err := pg.DB.Get(&temp, qNotificationInfoByOrgID, orgId); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}

	return &model.EmailNotificationInfo{
		Name:  temp.DisplayName,
		Email: temp.Email,
	}, nil
}

const qNotificationInfoByOrgID = `
SELECT  email, name
  FROM orgs
 WHERE id = $1`


// RotateSSOToken a new SSO secret.
func (pg *PGClient) RotateSSOSecret(secret, orgID string) error {
	result, err := pg.DB.Exec(qRotateSSOSecret, secret, orgID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrUserNotFound
	}
	return  nil
}

const qRotateSSOSecret = `
UPDATE orgs 
SET sso_secret = $1
WHERE id = $2 `


// GetSSOSecretByOrgID looks up an sso secret by its organisation ID or slug.
func (pg *PGClient) GetSSOSecretByOrgIDOrSlug(id string) (*messages.SSOSecret, error) {
	sso := &messages.SSOSecret{}

	if	err := pg.DB.Get(sso, qGetSSOSecret, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return sso, nil
}

const qGetSSOSecret = ` SELECT sso_secret FROM orgs WHERE id = $1 OR slug = $1 `

// DeleteSSOSecret deletes the SSO secret for an org.
func (pg *PGClient) DeleteSSOSecret(id string) error {
	result, err := pg.DB.Exec(qDeleteSSOSecret, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrUserNotFound
	}
	return nil
}

const qDeleteSSOSecret = `
UPDATE orgs 
SET sso_secret = NULL
WHERE id = $1`
