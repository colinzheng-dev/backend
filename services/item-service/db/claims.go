package db

import (
	"database/sql"
	"errors"
	"strings"

	// Import Postgres DB driver.

	"github.com/veganbase/backend/chassis"
	"github.com/veganbase/backend/services/item-service/model"
	"github.com/veganbase/backend/services/item-service/model/types"
)

// CreateClaim creates a new ownership claim.
func (pg *PGClient) CreateClaim(claim *model.Claim) error {
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

	// Generate new claim ID.
	claim.ID = chassis.NewID("claim")

	// Insert the claim.
	rows, err := tx.NamedQuery(createClaim, claim)
	if err != nil {
		return err
	}
	if !rows.Next() {
		return errors.New("couldn't create ownership claim")
	}
	err = rows.Scan(&claim.CreatedAt)
	return err
}

const createClaim = `
INSERT INTO
  ownership_claims (id, owner_id, item_id)
 VALUES (:id, :owner_id, :item_id)
 RETURNING created_at`

// ClaimByID looks up an ownership claim.
func (pg *PGClient) ClaimByID(id string) (*model.Claim, error) {
	claim := model.Claim{}
	err := pg.DB.Get(&claim, claimByID, id)
	if err != nil {
		return nil, err
	}
	return &claim, nil
}

const claimByID = `
SELECT id, owner_id, item_id, status, created_at
  FROM ownership_claims WHERE id = $1`

// Claims lists outstanding ownership claims, optionally filtering by
// user or claim status, with pagination.
func (pg *PGClient) Claims(allowedOwner []string, status *types.ApprovalState,
	page, perPage uint) ([]model.Claim, *uint, error) {
	results := []model.Claim{}
	var total uint
	q := getClaims + claimParamsWhere(allowedOwner, status) +
		` ORDER BY created_at DESC ` + chassis.Paginate(page, perPage)

	if err := pg.DB.Select(&results, q); err != nil {
		return nil, nil, err
	}
	if err := chassis.GetTotalResultsFromQuery(q, &total, pg.DB); err != nil {
		return nil, nil, err
	}
	return results, &total, nil
}

const getClaims = `
SELECT id, owner_id, item_id, status, created_at
  FROM ownership_claims WHERE `

func claimParamsWhere(allowedOwner []string, status *types.ApprovalState) string {
	es := []string{}
	if status != nil {
		es = append(es, `status = '`+status.String()+`'`)
	}
	if allowedOwner != nil && len(allowedOwner) != 0 {
		es = append(es, `owner_id IN ('`+strings.Join(allowedOwner, "', '")+`')`)
	}
	if len(es) == 0 {
		return `TRUE`
	}
	return strings.Join(es, ` AND `)
}

// DeleteClaim deletes an ownership claim.
func (pg *PGClient) DeleteClaim(id string, allowedOwners []string) error {
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

	// Retrieve the claim to check ownership.
	check := &model.Claim{}
	err = tx.Get(check, claimByID, id)
	if err == sql.ErrNoRows {
		return ErrClaimNotFound
	}
	if err != nil {
		return err
	}

	// Check owner.
	if len(allowedOwners) > 0 {
		allowed := false
		for _, o := range allowedOwners {
			if check.OwnerID == o {
				allowed = true
				break
			}
		}
		if !allowed {
			return ErrClaimNotOwned
		}
	}

	// Try to delete the claim.
	result, err := tx.Exec(deleteClaim, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrClaimNotFound
	}
	return err
}

const deleteClaim = `DELETE FROM ownership_claims WHERE id = $1`

// UpdateClaim updates the status of an ownership claim.
func (pg *PGClient) UpdateClaim(claim *model.Claim) error {
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

	check := &model.Claim{}
	err = tx.Get(check, getClaims+`id = $1`, claim.ID)
	if err == sql.ErrNoRows {
		return ErrClaimNotFound
	}
	if err != nil {
		return err
	}

	// Check read-only fields.
	if claim.OwnerID != check.OwnerID || claim.ItemID != check.ItemID ||
		claim.CreatedAt != check.CreatedAt {
		return ErrReadOnlyField
	}

	// Do the update.
	result, err := tx.Exec(updateClaim, claim.ID, claim.Status)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return ErrClaimNotFound
	}

	return err
}

const updateClaim = `UPDATE ownership_claims SET status = $2 WHERE id = $1`
