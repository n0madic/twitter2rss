package main

import (
	"log"
	"net/http"
	"os"

	"github.com/n0madic/twitter2rss"
)

func main() {
	http.HandleFunc("/", twitter2rss.HTTPHandler)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "//abs.twimg.com/favicons/twitter.ico", 301)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
