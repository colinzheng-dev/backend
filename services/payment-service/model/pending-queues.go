package model

import "time"

type PendingEvent struct {
	EventID    string     `db:"event_id"`
	IntentID   string     `db:"intent_id"`
	Reason     string     `db:"reason"`
	Attempts   int        `db:"attempts"`
	LastUpdate *time.Time `db:"last_update"`
	CreatedAt  time.Time  `db:"created_at"`
}

type PendingTransfer struct {
	ID                   int64     `db:"id"`
	Origin               string    `db:"origin"`
	Destination          string    `db:"destination"`
	Currency             string    `db:"currency"`
	SourceTransaction    string    `db:"source_transaction"`
	TotalValue           int       `db:"total_value"`
	FeeValue             int       `db:"fee_value"`
	TransferredValue     int       `db:"transferred_value"`
	FeeRemainder         float64   `db:"fee_remainder"`
	TransferredRemainder float64   `db:"transferred_remainder"`
	Reason               string    `db:"reason"`
	CreatedAt            time.Time `db:"created_at"`
}
