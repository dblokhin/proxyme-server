package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// runMetricsServers exposes metric server at /metrics endpoint
// using `METRICS_LISTEN_ADDR` (if it is not specified, metrics server is
// disabled)
func runMetrics(addr string) *http.Server {
	if len(addr) == 0 {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	const maxTimeout = 3 * time.Second
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       maxTimeout,
		ReadHeaderTimeout: maxTimeout,
		WriteTimeout:      maxTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("metrics http serve:", err)
		}
	}()

	return srv
}

func stopMetrics(srv *http.Server) error {
	if srv == nil {
		return nil
	}

	const shutdownTimeout = time.Second

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return srv.Shutdown(ctx)
}
