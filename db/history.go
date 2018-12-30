package db

import "time"

// History defines the schema for the `history` table
type History struct {
	LastPubdate *time.Time
	FeedID      int
}
