package main

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/cryptopay-dev/go-metrics"
	nats "github.com/nats-io/go-nats"
)

func main() {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	err = metrics.Setup("nats://localhost:4222", "my-metrics", hostname, nats.Timeout(time.Minute), nats.UserInfo("user", "password"))
	if err != nil {
		log.Fatal(err)
	}

	go metrics.Watch(time.Second * 10)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer track(time.Now(), r.RemoteAddr)

		time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
		w.Write([]byte(`Hello, friend`))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func track(s time.Time, ip string) {
	elapsed := time.Since(s)
	err := metrics.SendWithTagsAndWait(metrics.M{
		"elapsed": elapsed,
	}, metrics.T{
		"ip": ip,
	}, "request", "index")

	if err != nil {
		log.Fatal(err)
	}
}
