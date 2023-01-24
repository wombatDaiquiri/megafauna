package main

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/mattn/go-mastodon"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wombatDaiquiri/megafauna/database"
	"gorm.io/gorm"
	"net/http"
	"time"
)

// TODO: separate commands for sub-functions
func main() {
	// ctx := context.Background()
	// ctx, cancel := context.WithCancel(ctx)

	// fmt.Println("connecting to database...")
	// db, err := gorm.Open(sqlite.Open("local.db"), &gorm.Config{})
	// if err != nil {
	// 	panic(err)
	// }
	//db = db.Debug()
	// err = db.AutoMigrate(&database.Status{})
	// if err != nil {
	// 	panic(err)
	// }
	// err = db.AutoMigrate(&database.Mention{})
	// if err != nil {
	// 	panic(err)
	// }
	// err = db.AutoMigrate(&database.Tag{})
	// if err != nil {
	// 	panic(err)
	// }

	// sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// go func() {
	// 	sig := <-sigs
	// 	fmt.Printf("\nreceived signal %v, cancelling context\n", sig.String())
	// 	cancel()
	// }()

	//servers := []string{
	//	"https://mastodon.social/",
	//}
	//var wg sync.WaitGroup
	//fmt.Println("engaging thrusters")
	//for _, server := range servers {
	//	server := server
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		scrapePublicStream(ctx, db, server)
	//	}()
	//}

	fmt.Println("setup completed, emiting metrics")
	router := chi.NewRouter()
	router.Get("/metrics", promhttp.Handler().ServeHTTP)
	if err := http.ListenAndServe(":3000", router); err != nil {
		panic(err)
	}
	//wg.Wait()
	fmt.Println("exiting")
}

func scrapePublicStream(ctx context.Context, db *gorm.DB, server string) {
	for {
		fmt.Println("starting stream for", server)
		select {
		case <-ctx.Done():
			return
		default:
		}
		cli := mastodon.NewClient(&mastodon.Config{
			Server: server,
		})
		stream, err := cli.StreamingPublic(ctx, false)
		if err != nil {
			panic(err)
		}
		events := 0
		timer := time.Now()
		var handleDuration time.Duration
		var contentSize int64
		for rawEvent := range stream {
			select {
			case <-ctx.Done():
				return
			default:
			}

			handleStart := time.Now()
			events++
			if events%100 == 0 {
				fmt.Print(".")
			}
			if events%1000 == 0 {
				gatheringTime := time.Now().Sub(timer)
				fmt.Printf(`
[%s] we have just scraped %v events! (stats for last 1000 events)
- collected in %v (avg. %v events/hour)
- handled in %v (avg. %v/event)
- content size (avg / post): %v
`, server, events, gatheringTime, 1000*3600/gatheringTime.Seconds(), handleDuration, handleDuration/1000, ByteCountSI(contentSize/1000))

				handleDuration = 0
				contentSize = 0
				timer = time.Now()
			}

			if event, ok := rawEvent.(*mastodon.UpdateEvent); ok && event != nil && event.Status != nil {
				status := database.StatusFromMastodonEvent(*event.Status)
				err := db.Create(&status).Error
				if err != nil {
					fmt.Println("\n[DB EVENT ERR]", err)
					handleDuration += time.Now().Sub(handleStart)
					continue
				}
			}
			handleDuration += time.Now().Sub(handleStart)
		}
		fmt.Println("stream broken for", server)
	}
}

// source: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
