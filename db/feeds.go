package db

// Feed defines the schema for the `feeds` table
type Feed struct {
	ID      int
	FeedURL string
}
