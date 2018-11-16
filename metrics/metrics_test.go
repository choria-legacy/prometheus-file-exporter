package metrics

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics")
}

var _ = Describe("Metrics", func() {
	var (
		log  *logrus.Entry
		path string
		err  error
	)

	BeforeEach(func() {
		logger := logrus.New()
		logger.SetOutput(ioutil.Discard)
		log = logrus.NewEntry(logger)
		path, err = filepath.Abs("testdata")
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("New", func() {
		It("Should initialize from disk and set up a watcher", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			metrics, err := New(ctx, path, true, log)
			Expect(err).ToNot(HaveOccurred())
			Expect(metrics.metrics).To(HaveKey(filepath.Join(path, "valid.json")))
			Expect(metrics.metrics).ToNot(HaveKey(filepath.Join(path, "test_test.json")))

			metric := NewMetric("test_test", map[string]string{})
			metric.Save(path)
			tfile := filepath.Join(path, "test_test.json")
			defer os.Remove(tfile)

			time.Sleep(20 * time.Millisecond)
			Expect(metrics.metrics).To(HaveKey(tfile))

			metrics.metrics[tfile].Incr(20)

			os.Remove(tfile)
			time.Sleep(20 * time.Millisecond)
			Expect(metrics.metrics).ToNot(HaveKey(tfile))
		})

		It("Should save and reload from disk on save", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			metrics, err := New(ctx, path, true, log)
			Expect(err).ToNot(HaveOccurred())

			metric := NewMetric("test_test", map[string]string{})
			metric.Set(10)
			metric.Save(path)
			tfile := filepath.Join(path, "test_test.json")
			defer os.Remove(tfile)

			time.Sleep(20 * time.Millisecond)

			Expect(metrics.metrics[tfile].Value).To(Equal(float64(10)))

			metric.Set(20)
			err = metric.Save(path)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(20 * time.Millisecond)
			Expect(metrics.metrics[tfile].Value).To(Equal(float64(20)))
		})
	})
})
