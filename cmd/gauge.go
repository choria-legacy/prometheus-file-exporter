package cmd

import (
	"github.com/choria-io/prometheus-file-exporter/metrics"
)

func gauge() error {
	metric := metrics.NewMetric(metricName, map[string]string{})

	metric.Set(value)
	return metric.Save(path)
}
