package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/araddon/dateparse"
	_ "github.com/mattn/go-sqlite3" // Sqlite
	"github.com/n0ncetonic/drgero-feed-monitor/feed"
)

// AddFeed adds an RSS/Atom feed to the feed table
func AddFeed(db *sql.DB, url string) error {
	stmt, err := db.Prepare("INSERT INTO feeds(feed_url) VALUES(?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Stmt(stmt).Exec(url)
	if err != nil {
		log.Println("rolling back transaction")
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// GetFeeds gets all feeds
func GetFeeds(db *sql.DB) (feeds feed.Feeds, err error) {
	stmt, err := db.Prepare("SELECT id,feed_url FROM feeds")
	if err != nil {
		return feeds, err
	}
	defer stmt.Close()

	tx, err := db.Begin()
	if err != nil {
		return feeds, err
	}

	rows, err := tx.Stmt(stmt).Query()
	if err != nil {
		return feeds, err
	}

	for rows.Next() {
		var feed feed.Feed
		var id string
		var url string
		err := rows.Scan(&id, &url)
		if err != nil {
			log.Printf("unable to get information about feed: %v", err)
		}
		feed.ID = id
		feed.URL = url
		feeds.Feeds = append(feeds.Feeds, feed)

	}
	rows.Close()
	tx.Commit()
	return feeds, nil
}

// GetFeedIDFromURL gets the ID of a feed based on a URL
func GetFeedIDFromURL(db *sql.DB, url string) (id int, err error) {
	stmt, err := db.Prepare("SELECT id FROM feeds where feed_url='?'")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(url)

	err = row.Scan(&id)
	if err != nil {
		log.Printf("failed to get ID for %s : %v", url, err)
		return 0, nil
	}

	return id, nil
}

// GetPubDate gets a stored pubdate for a feed
func GetPubDate(db *sql.DB, id string) (ts time.Time, err error) {
	stmt, err := db.Prepare("SELECT last_pubdate FROM history where feed_id=?")
	if err != nil {
		return time.Time{}, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(id)

	var pubdate string

	row.Scan(&pubdate)

	if len(pubdate) < 1 {
		log.Printf("no previous date found so setting to 1 day ago")
		return time.Now().AddDate(0, 0, -1), nil
	}

	ts, err = dateparse.ParseAny(pubdate)
	if err != nil {
		log.Printf("failed to parse pubdate format: %v", err)
		return time.Time{}, err
	}

	return ts, nil
}

// UpdateHistory updates the last_pubdate row for a feed_id in the `history` table
func UpdateHistory(db *sql.DB, feedID string, pb string) error {
	stmt, err := db.Prepare("REPLACE INTO history(last_pubdate,feed_id) VALUES(?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(pb, feedID)
	if err != nil {
		return err
	}
	return nil
}
