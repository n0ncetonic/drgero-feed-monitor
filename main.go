package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	_ "github.com/mattn/go-sqlite3" // Sqlite
	"github.com/n0ncetonic/drgero-feed-monitor/config"
	fdb "github.com/n0ncetonic/drgero-feed-monitor/db"
	"github.com/n0ncetonic/drgero-feed-monitor/event"
)

type options struct {
	Add      *string `short:"a" long:"add" description:"Adds a feed to the database"`
	Interval int     `short:"i" long:"interval" description:"Time in minutes to wait before checking feeds" default:"5"`
}

var opts options
var parser = flags.NewParser(&opts, flags.Default)
var cfg config.Cfg

func makeTables(db *sql.DB) error {
	log.Println("Creating `feeds` table if it does not exist")

	stmt, err := db.Prepare(
		"CREATE TABLE IF NOT EXISTS feeds (id INTEGER PRIMARY KEY, feed_url VARCHAR(255) NOT NULL UNIQUE)",
	)
	if err != nil {
		log.Printf("failed to prepare create table statement for `feeds` table: %v ", err)
	}
	defer stmt.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
	}

	_, err = tx.Stmt(stmt).Exec()
	if err != nil {
		log.Fatalf("failed to create `feeds` table: %v", err)
		tx.Rollback()
	}
	tx.Commit()

	log.Println("Creating `history` table if it does not exist")
	stmt, err = db.Prepare(
		"CREATE TABLE IF NOT EXISTS history (last_pubdate TEXT, feed_id TEXT UNIQUE)",
	)
	if err != nil {
		log.Printf("failed to prepare create table statement for `history` table: %v ", err)
	}

	tx, err = db.Begin()
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
	}

	_, err = tx.Stmt(stmt).Exec()
	if err != nil {
		log.Fatalf("failed to create `history` table: %v", err)
		tx.Rollback()
	}

	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func open() (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", "./feeds.db")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func checkFeeds(db *sql.DB) {
	feeds, err := fdb.GetFeeds(db)
	if err != nil {
		log.Println(err)
	}

	for _, feed := range feeds.Feeds {
		lpd, err := fdb.GetPubDate(db, feed.ID)
		if err != nil {
			log.Printf("failed to get last pubdate for feed ID `%s` : %v", feed.ID, err)
		}

		items, mostRecent, err := feed.Parse(&lpd)
		if err != nil {
			log.Printf("failed to get items for feed `%s`: %v", feed.URL, err)
			break
		}

		err = fdb.UpdateHistory(db, feed.ID, mostRecent)
		if err != nil {
			log.Printf("%v", err)
		}
		for _, item := range items.Items {
			var e event.Event

			e.Link = item.Link
			e.Title = item.Title

			err := e.Send(cfg.Host)
			if err != nil {
				log.Printf("failed to send event %v : %v", e, err)
			}
		}
	}

}

func init() {
	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}

func main() {
	db, err := open()
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	err = makeTables(db)
	if err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	if opts.Add != nil && *opts.Add != "" {
		err = fdb.AddFeed(db, *opts.Add)
		if err != nil {
			log.Fatalf("Failed to add %s as feed: %v", *opts.Add, err)
		}
		log.Printf("Added %s to feeds", *opts.Add)
		return
	}

	err = cfg.Read("config.json")
	if err != nil {
		log.Fatalf("Failed to read config.json: %v", err)
		return
	}

	ticker := time.NewTicker(time.Duration(opts.Interval) * time.Minute).C

	log.Println("Starting Feed Monitor")
	for {
		select {
		case <-ticker:
			log.Println("Checking feeds for updates")
			checkFeeds(db)
		}
	}
}
