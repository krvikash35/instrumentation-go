package main

import (
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"contrib.go.opencensus.io/exporter/ocagent"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

func main() {
	// Firstly, we'll register ochttp Server views.
	if err := view.Register(ochttp.DefaultServerViews...); err != nil {
		log.Fatalf("Failed to register server views for HTTP metrics: %v", err)
	}

	// The handler containing your business logic to process requests.
	originalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("inside")
		// Consume the request's body entirely.
		io.Copy(ioutil.Discard, r.Body)

		// Generate some payload of random length.
		res := strings.Repeat("a", rand.Intn(99971)+1)

		// Sleep for a random time to simulate a real server's operation.
		time.Sleep(time.Duration(rand.Intn(977)+1) * time.Millisecond)

		// Finally write the body to the response.
		w.Write([]byte("Hello, World! " + res))
	})
	och := &ochttp.Handler{
		Handler: originalHandler, // The handler you'd have used originally
	}

	// Enable observability to extract and examine stats.
	enableObservabilityAndExporters(och)
}

func enableObservabilityAndExporters(h http.Handler) {

	oce, err := ocagent.NewExporter(ocagent.WithAddress("localhost:55678"), ocagent.WithInsecure(), ocagent.WithServiceName("instr-go-demo-serivce"), ocagent.WithReconnectionPeriod(5*time.Second))

	// Stats exporter: Prometheus
	// pe, err := prometheus.NewExporter(prometheus.Options{
	// 	Namespace: "ochttp_tutorial",
	// })
	if err != nil {
		log.Fatalf("Failed to create the stats exporter: %v", err)
	}

	// agentEndpointURI := "localhost:6831"
	// collectorEndpointURI := "http://localhost:14268/api/traces"
	// je, err := jaeger.NewExporter(jaeger.Options{
	// 	AgentEndpoint:     agentEndpointURI,
	// 	CollectorEndpoint: collectorEndpointURI,
	// 	ServiceName:       "demo",
	// })

	view.RegisterExporter(oce)
	trace.RegisterExporter(oce)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	mux := http.NewServeMux()
	// mux.Handle("/metrics", pe)
	mux.Handle("/hello", h)
	log.Fatal(http.ListenAndServe(":7777", mux))
}
