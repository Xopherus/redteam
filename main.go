package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const gzip = ""

func main() {
	// register prometheus metrics
	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "promhttp_metric_handler_request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"handler", "method"},
	)
	prometheus.MustRegister(duration)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// 33% of time, randomly return 5xxs
		if rand.Intn(3) == 0 {
			log.Printf("returning 5XX")
			w.WriteHeader(502)
			return
		}

		// 50% of time sometimes sleep
		if rand.Intn(2) == 0 {
			log.Printf("sleeping")
			time.Sleep(time.Duration(10 + rand.Intn(20)))
		}

		// add trailing headers which will force entire body to be read
		w.Header().Set("Trailer", "AtEnd1, AtEnd2")
		w.Header().Add("Trailer", "AtEnd3")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8") // normal header
		w.WriteHeader(307)

		w.Header().Set("AtEnd1", "value 1")

		b := make([]byte, 100)
		for i := 0; i < 100000; i++ {
			if _, err := rand.Read(b); err != nil {
				break
			}
			w.Write(b)
		}
		w.Header().Set("AtEnd2", "value 2")
		w.Header().Set("AtEnd3", "value 3")
	})

	http.HandleFunc("/", promhttp.InstrumentHandlerDuration(
		duration.MustCurryWith(prometheus.Labels{"handler": "/"}),
		handler,
	))

	http.Handle("/status", promhttp.Handler())

	s := &http.Server{Addr: ":8080"}
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("server closed with err %s", err)
	}
}
