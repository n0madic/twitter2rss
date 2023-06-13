package main

import (
	_ "embed"
	"log"
	"net/http"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/n0madic/twitter2rss"
	limiter "github.com/ulule/limiter/v3"
	mhttp "github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	mstore "github.com/ulule/limiter/v3/drivers/store/memory"
	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

type config struct {
	Port            string        `env:"PORT" envDefault:"8000"`
	CacheCapacity   int           `env:"CACHE_CAPACITY" envDefault:"10000"`
	CacheTTL        time.Duration `env:"CACHE_TTL" envDefault:"15m"`
	RateLimit       string        `env:"RATE_LIMIT" envDefault:"1-M"`
	RateLimitHeader string        `env:"RATE_LIMIT_HEADER" envDefault:"X-Forwarded-For"`
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

	rate, err := limiter.NewRateFromFormatted(cfg.RateLimit)
	if err != nil {
		log.Fatal(err)
	}
	store := mstore.NewStore()
	lmt := mhttp.NewMiddleware(limiter.New(store, rate, limiter.WithClientIPHeader(cfg.RateLimitHeader)))

	http.Handle("/",
		cacheClient.Middleware(
			lmt.Handler(
				http.HandlerFunc(twitter2rss.HTTPHandler),
			),
		),
	)

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "//abs.twimg.com/favicons/twitter.ico", http.StatusMovedPermanently)
	})

	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
