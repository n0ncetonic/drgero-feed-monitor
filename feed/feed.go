package feed

import (
	"log"
	"time"

	"github.com/mmcdole/gofeed"
)

// Feed holds information for fetching an RSS or Atom feed
type Feed struct {
	ID  string
	URL string
}

// Feeds holds multiple `Feed` objects
type Feeds struct {
	Feeds []Feed
}

// Item is a single item in a feed
type Item struct {
	Title string
	Link  string
}

// Items is a collection of `Item`
type Items struct {
	Items []Item
}

// Parse parses a feed
func (f *Feed) Parse(lpd *time.Time) (items Items, mostRecent string, err error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(f.URL)
	if err != nil {
		log.Printf("failed to parse feed: %v", err)
		return Items{}, "", err
	}

	latest := lpd

	for _, item := range feed.Items {
		var entry Item
		var pub *time.Time

		if item.Published != "" {
			pub = item.PublishedParsed
		} else if item.Updated != "" {
			pub = item.UpdatedParsed
		} else {
			break
		}

		// if item.PublishedParsed.Before(*lpd) || item.PublishedParsed.Equal(*lpd) || item.UpdatedParsed.Before(*lpd) || item.UpdatedParsed.Equal(*lpd) {
		// 	break
		// }

		if pub.Before(*lpd) || pub.Equal(*lpd) {
			break
		}

		if pub.After(*latest) {
			latest = pub
		}

		entry.Title = item.Title
		entry.Link = item.Link
		items.Items = append(items.Items, entry)
	}
	return items, latest.String(), nil
}
