package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/n0madic/twitter2rss"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

func main() {
	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000),
	)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(10*time.Minute),
	)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", cacheClient.Middleware(http.HandlerFunc(twitter2rss.HTTPHandler)))
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "//abs.twimg.com/favicons/twitter.ico", 301)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
