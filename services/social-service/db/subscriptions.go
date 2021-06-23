package db

const qCreateUserSubscription = `
INSERT INTO
	subscriptions (user_id, subscription_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`

// CreateUserSubscription associates the given userID with a subscriptionID
func (pg *PGClient) CreateUserSubscription(userID, subscriptionID string) error {
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

	_, err = tx.Exec(qCreateUserSubscription, userID, subscriptionID)

	return err
}

const qListUserSubscriptions = `
SELECT subscription_id
FROM subscriptions
WHERE user_id = $1
`

// ListUserSubscriptions returns a list of IDs being followed by a user
func (pg *PGClient) ListUserSubscriptions(userID string) ([]string, error) {
	subs := []string{}
	err := pg.DB.Select(&subs, qListUserSubscriptions, userID)
	return subs, err
}

const qDeleteUserSubscription = `
DELETE FROM subscriptions
WHERE user_id = $1 AND subscription_id = $2
`

// DeleteUserSubscription removes a subscriptionID from the userID
func (pg *PGClient) DeleteUserSubscription(userID, subscriptionID string) error {
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

	_, err = tx.Exec(qDeleteUserSubscription, userID, subscriptionID)

	return err
}


const qListFollowers = `
SELECT user_id
FROM subscriptions
WHERE subscription_id = $1
`

// ListFollowers returns a list of userIDs currently following subscriptionID
func (pg *PGClient) ListFollowers(subscriptionID string) ([]string, error) {
	followers := []string{}
	err := pg.DB.Select(&followers, qListFollowers, subscriptionID)
	return followers, err
}
