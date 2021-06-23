package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"

	// Import Postgres DB driver.
	_ "github.com/lib/pq"

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/user-service/model"
)

// UserByID looks up a user by their user ID.
func (pg *PGClient) UserByID(id string) (*model.User, error) {
	user := &model.User{}
	err := pg.DB.Get(user, userBy + "id = $1", id)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

const userBy = `
SELECT id, email, name, display_name, avatar,
       country, is_admin, last_login, api_key
  FROM users
 WHERE `

const userWithSecretBy = `
SELECT id, email, name, display_name, avatar,
       country, is_admin, last_login, api_key, secret_key
  FROM users
 WHERE `


// UserByAPIKey looks up a user by it's API key.
func (pg *PGClient) UserByAPIKey(apiKey string) (*model.User, error) {
	user := &model.User{}

	if err := pg.DB.Get(user, userWithSecretBy + "api_key = $1", apiKey); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}


// UsersByIDs returns the full user model for a given list of user
// IDs as a map indexed by the ID.
func (pg *PGClient) UsersByIDs(ids []string) (map[string]*model.User, error) {
	if len(ids) == 0 {
		return map[string]*model.User{}, nil
	}

	query, args, err := sqlx.In(usersByIDs, ids)
	if err != nil {
		return nil, err
	}

	query = pg.DB.Rebind(query)

	users := []*model.User{}
	err = pg.DB.Select(&users, query, args...)
	if err != nil {
		return nil, err
	}

	result := map[string]*model.User{}
	for _, u := range users {
		result[u.ID] = u
	}

	return result, nil
}

const usersByIDs = `
SELECT id, email, name, display_name, avatar,
       country, is_admin, last_login, api_key
  FROM users
 WHERE id IN (?)`

// Users gets a list of user in reverse last login date order,
// optionally filtered by a search term and paginated.
func (pg *PGClient) Users(search string, page, perPage uint) ([]model.User, error) {
	var err error
	results := []model.User{}
	if search != "" {
		err = pg.DB.Select(&results,
			userListWithSearch+chassis.Paginate(page, perPage), "%"+search+"%")
	} else {
		err = pg.DB.Select(&results,
			userList+chassis.Paginate(page, perPage))
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}

const userList = `
SELECT id, email, name, display_name, avatar,
       country, is_admin, last_login, api_key
  FROM users
 ORDER BY last_login DESC`

const userListWithSearch = `
SELECT id, email, name, display_name, avatar,
       country, is_admin, last_login, api_key
  FROM users
 WHERE email LIKE $1 OR name LIKE $1 OR display_name LIKE $1
 ORDER BY last_login DESC`

// LoginUser performs login actions for a given email address:
//
//  - If an account with the given email address already exists in the
//    database, then set the account's last_login field to the current time.
//
//  - If an account with the given email address does not already
//    exist in the database, then create a new user account with the
//    given email address, defaulting the name and display_name
//    fields to match the email address and setting the new
//    account's last_login field to the current time.
//
// In both cases, return the full user record of the logged in user.
func (pg *PGClient) LoginUser(email string, avatarGen func() string) (*model.User, bool, error) {
	tx, err := pg.DB.Beginx()
	if err != nil {
		return nil, false, err
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	user := &model.User{}
	newUser := false
	err = tx.Get(user, userBy + "email = $1 ", email)
	log.Info().Msg(fmt.Sprintf("db.LoginUser: err = %v", err))

	if err == nil {
		user.LastLogin = time.Now()
		_, err = tx.Exec(updateLastLogin, user.LastLogin, user.ID)
	} else {
		log.Info().Msg("NEW USER")
		userID := chassis.NewID("usr")
		newUser = true
		avatar := avatarGen()
		user = &model.User{
			ID:          userID,
			Email:       email,
			Name:        &email,
			DisplayName: &email,
			Avatar:      &avatar,
			IsAdmin:     false,
			LastLogin:   time.Now(),
		}
		_, err = tx.NamedExec(createUser, user)
	}

	if err != nil {
		return nil, false, err
	}
	return user, newUser, nil
}

const updateLastLogin = `UPDATE users SET last_login = $1 WHERE id = $2`

const createUser = `
INSERT INTO users (id, email, name, display_name, avatar, is_admin, last_login)
     VALUES (:id, :email, :name, :display_name, :avatar, :is_admin, :last_login)`

// UpdateUser updates the user's details in the database. The id,
// email, last_login and api_key fields are read-only using this
// method.
func (pg *PGClient) UpdateUser(user *model.User) error {
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

	check := &model.User{}
	err = tx.Get(check, userBy + "id = $1", user.ID)
	if err == sql.ErrNoRows {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}

	// Check read-only fields. NOTE: only administrators can update the
	// is_admin field. And they can't switch their own is_admin flag off
	// themselves. This needs to be checked in the handler that calls
	// the update, since we don't have the authentication information
	// needed to make this decision here.
	if user.Email != check.Email ||
		user.LastLogin != check.LastLogin || user.APIKey != check.APIKey {
		return ErrReadOnlyField
	}

	// Fix other field values: name fields should use empty strings for
	// null values.
	empty := ""
	if user.Name == nil {
		user.Name = &empty
	}
	if user.DisplayName == nil {
		user.DisplayName = &empty
	}

	// Do the update.
	_, err = tx.NamedExec(updateUser, user)
	if err != nil {
		return err
	}
	return nil
}

const updateUser = `
UPDATE users SET name=:name, display_name=:display_name,
                 avatar=:avatar, country=:country, is_admin=:is_admin
 WHERE id = :id`

// DeleteUser deletes the given user account.
// TODO: ADD SOME SORT OF ARCHIVAL MECHANISM INSTEAD.
func (pg *PGClient) DeleteUser(id string) error {
	result, err := pg.DB.Exec(deleteUser, id)
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

const deleteUser = "DELETE FROM users WHERE id = $1"

// SaveHashedAPIKey saves the hashed api key.
func (pg *PGClient) SaveHashedAPIKey(key, secret, userID string) error {
	result, err := pg.DB.Exec(rotateAPIKey, key, secret, userID)
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

const rotateAPIKey = `
UPDATE users 
SET api_key = $1,
 	secret_key = $2 
WHERE id = $3`

// DeleteAPIKey deletes the API key for a user.
func (pg *PGClient) DeleteAPIKey(id string) error {
	result, err := pg.DB.Exec(deleteAPIKey, id)
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

const deleteAPIKey = `
UPDATE users 
SET api_key = NULL, 
    secret_key = NULL 
WHERE id = $1`


func (pg *PGClient) NotificationInfoByUserId(userId string) (*model.EmailNotificationInfo, error) {
	info := &model.EmailNotificationInfo{}
	if err := pg.DB.Get(info, qNotificationInfoByUserID, userId); err != nil {
	 	if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return info, nil
}

const qNotificationInfoByUserID = `
SELECT  email, display_name
  FROM users
 WHERE id = $1`