package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/choria-io/prometheus-file-exporter/metrics"
)

func list() error {
	m, err := metrics.New(ctx, path, false, log)
	if err != nil {
		return fmt.Errorf("could not set up metrics: %s", err)
	}

	fmt.Printf("Listing metrics found in %s\n\n", path)

	for n, metric := range m.Metrics() {
		if filter != "" {
			if !strings.Contains(metric.Name, filter) {
				continue
			}
		}

		fmt.Printf("%s (%s)\n", metric.Name, n)
		fmt.Printf("     Type: %s\n", metric.Type)
		fmt.Printf("  Updated: %s\n", time.Unix(0, metric.Timestamp).Format(time.RFC3339))
		fmt.Printf("    Value: %v\n", metric.Value)
		fmt.Println()
	}

	return nil
}
