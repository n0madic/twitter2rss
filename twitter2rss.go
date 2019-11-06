package twitter2rss

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/feeds"
	twitterscraper "github.com/n0madic/twitter-scraper"
)

// Index html for empty requests
var Index string

// Twitter2RSS return RSS from twitter timeline
func Twitter2RSS(screenName string, pages int, excludeReplies bool) (string, error) {
	feed := &feeds.Feed{
		Title:       "Twitter feed @" + screenName,
		Link:        &feeds.Link{Href: "https://twitter.com/" + screenName},
		Description: "Twitter feed @" + screenName + " through Twitter to RSS proxy by Nomadic",
	}

	for tweet := range twitterscraper.GetTweets(screenName, pages) {
		if tweet.Error != nil {
			return "", tweet.Error
		}

		if excludeReplies && tweet.IsRetweet {
			continue
		}

		if tweet.TimeParsed.After(feed.Created) {
			feed.Created = tweet.TimeParsed
		}

		var title string

		titleSplit := strings.FieldsFunc(tweet.Text, func(r rune) bool {
			return r == '\n' || r == '!' || r == '?' || r == ':' || r == '<' || r == '.' || r == ','
		})
		if len(titleSplit) > 0 {
			if strings.HasPrefix(titleSplit[0], "a href") || strings.HasPrefix(titleSplit[0], "http") {
				title = "link"
			} else {
				title = titleSplit[0]
			}
		}

		for _, img := range tweet.Photos {
			tweet.HTML += fmt.Sprintf("\n<img src=\"%s\">", img)

		}
		for _, video := range tweet.Videos {
			tweet.HTML += fmt.Sprintf("\n<img src=\"%s\">", video.Preview)
		}

		tweet.HTML = strings.Replace(tweet.HTML, "\n", "<br>", -1)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(tweet.HTML))
		if err == nil {
			doc.Find("a.twitter-timeline-link").Each(func(i int, sel *goquery.Selection) {
				if a, exists := sel.Attr("data-expanded-url"); exists {
					u, err := url.Parse(a)
					if err == nil && u.IsAbs() {
						sel.SetAttr("href", a)
					}
				}
			})
			if html, err := doc.Html(); err == nil {
				tweet.HTML = html
			}
		}

		feed.Add(&feeds.Item{
			Author:      &feeds.Author{Name: screenName},
			Created:     tweet.TimeParsed,
			Description: tweet.HTML,
			Id:          tweet.PermanentURL,
			Link:        &feeds.Link{Href: tweet.PermanentURL},
			Title:       title,
		})
	}

	return feed.ToRss()
}

// HTTPHandler function
func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = strings.TrimPrefix(html.EscapeString(r.URL.Path), "/")
	}
	if name != "" {
		pageCount, _ := strconv.Atoi(r.URL.Query().Get("pages"))
		statusCount, _ := strconv.Atoi(r.URL.Query().Get("count"))
		if statusCount > 0 && pageCount == 0 {
			pageCount = statusCount / 10
		}
		if pageCount < 1 {
			pageCount = 1
		} else if pageCount > 10 {
			pageCount = 10
		}

		excludeReplies := r.URL.Query().Get("exclude_replies") == "on"

		log.Printf("Process timeline @%s (pages: %d, exclude_replies: %v)", name, pageCount, excludeReplies)
		rss, err := Twitter2RSS(name, pageCount, excludeReplies)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(rss))
	} else {
		w.Write([]byte(Index))
	}
}
