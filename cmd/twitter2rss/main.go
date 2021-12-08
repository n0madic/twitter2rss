package main

import (
	_ "embed"
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/n0madic/twitter2rss"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

type config struct {
	Port          string        `env:"PORT" envDefault:"8000"`
	CacheCapacity int           `env:"CACHE_CAPACITY" envDefault:"10000"`
	CacheTTL      time.Duration `env:"CACHE_TTL" envDefault:"15m"`
}

//go:embed index.html
var index string

func init() {
	twitter2rss.Index = index
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(cfg.CacheCapacity),
	)
	if err != nil {
		log.Fatal(err)
	}

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(cfg.CacheTTL),
	)
	if err != nil {
		log.Fatal(err)
	}

	lmt := tollbooth.NewLimiter(1, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})

	http.Handle("/",
		cacheClient.Middleware(
			tollbooth.LimitFuncHandler(lmt,
				http.HandlerFunc(twitter2rss.HTTPHandler),
			),
		),
	)

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "//abs.twimg.com/favicons/twitter.ico", http.StatusMovedPermanently)
	})

	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
