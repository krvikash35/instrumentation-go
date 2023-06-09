// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func initTracer() (*sdktrace.TracerProvider, error) {

	//stdouttrace
	//otlptrace(newrelic)
	//otlptrace(jaeger)
	//otlptrace(otlp)

	// Create stdout exporter to be able to retrieve
	// the collected spans.
	// exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	// exporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient(otlptracehttp.WithEndpoint("otlp.nr-data.net"), otlptracehttp.WithHeaders(map[string]string{"api-key": "d315a109547d9e69c95143935599d9cbe9deNRAL"})))
	exporter, err := otlptrace.New(context.Background(), otlptracehttp.NewClient(otlptracehttp.WithEndpoint("localhost:4318"), otlptracehttp.WithInsecure()))

	if err != nil {
		return nil, err
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName("ExampleService"))),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, err
}

func initMeter() (*sdkmetric.MeterProvider, error) {
	//stdoutmetric
	//otlpmetrichttp(newrelic)
	//otlpmetrichttp(otlp)
	//prometheus

	// exp, err := stdoutmetric.New()

	// exp, err := otlpmetrichttp.New(context.Background(), otlpmetrichttp.WithEndpoint("otlp.nr-data.net"), otlpmetrichttp.WithHeaders(map[string]string{"api-key": "d315a109547d9e69c95143935599d9cbe9deNRAL"}))
	exp, err := otlpmetrichttp.New(context.Background(), otlpmetrichttp.WithEndpoint("localhost:4318"), otlpmetrichttp.WithInsecure())
	// exp, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp, sdkmetric.WithInterval(3*time.Second))))
	// mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exp))
	global.SetMeterProvider(mp)
	return mp, nil
}

func main() {

	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	mp, err := initMeter()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := mp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down meter provider: %v", err)
		}
	}()

	uk := attribute.Key("username")

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		span := trace.SpanFromContext(ctx)
		bag := baggage.FromContext(ctx)
		span.AddEvent("handling this...", trace.WithAttributes(uk.String(bag.Member("username").Value())))

		_, _ = io.WriteString(w, "Hello, world!\n")
	}

	otelHandler := otelhttp.NewHandler(WithRouteTag("/hello", http.HandlerFunc(helloHandler)), "Hello")

	http.Handle("/hello", otelHandler)
	http.Handle("/metrics", promhttp.Handler())
	err = http.ListenAndServe(":7777", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// WithRouteTag is add custom function based on otelhttp.WithRouteTag
// this is to be used temporarily pending https://github.com/open-telemetry/opentelemetry-go-contrib/pull/615 to be merged
func WithRouteTag(route string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attr := semconv.HTTPRouteKey.String(route)

		span := trace.SpanFromContext(r.Context())
		span.SetAttributes(attr)

		labeler, _ := otelhttp.LabelerFromContext(r.Context())
		labeler.Add(attr)
		h.ServeHTTP(w, r)
	})
}
