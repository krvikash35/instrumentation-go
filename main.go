package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alexcesaro/statsd.v2"
)

const sd_metric = "sd_http_total_requests"
const prom_metric = "prom_http_total_requests"

func main() {
	// c := promauto.NewCounterVec(prometheus.CounterOpts{
	// 	Name: prom_metric,
	// }, []string{"path"})

	// sdClient := newStatsdClient()

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		log.Println("increment134")
		// w.WriteHeader(http.StatusOK)
		// w.Write([]byte("pong"))

		// c.WithLabelValues("/ping").Inc()

		// // pingStats := sdClient.Clone(statsd.Tags("path", "/ping"))
		// log.Println("increment")
		// sdClient.Increment(sd_metric)

	})

	http.ListenAndServe(":8081", nil)
}

func newStatsdClient() *statsd.Client {
	c, err := statsd.New(statsd.Address("statsd:8125"), statsd.ErrorHandler(func(err error) {
		log.Printf("failed to sent metric to statsd :%s", err)
	}))
	if err != nil {
		log.Printf("failed to connect to statsd :%s", err)
	}
	return c
}
