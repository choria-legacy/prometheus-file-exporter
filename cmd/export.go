package cmd

import (
	"fmt"
	"net/http"

	"github.com/choria-io/prometheus-file-exporter/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func export() error {
	_, err := metrics.New(ctx, path, true, log)
	if err != nil {
		return fmt.Errorf("could not set up metrics exporter: %s", err)
	}

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

	for {
		select {
		case <-ctx.Done():
			return nil
		}

	}
}
