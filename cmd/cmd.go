package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	path    string
	logfile string
	debug   bool
	ctx     context.Context
	cancel  func()
	log     *logrus.Entry
	port    int
	pidfile string
	err     error
	app     *kingpin.Application
	filter  string

	metricName string
	value      float64

	DefaultPath string
	Version     string
	Sha         string
)

func Run() {
	app = kingpin.New("pfe", "The Choria Prometheus File Exporter")
	app.Author("R.I.Pienaar <rip@devco.net>")
	app.Version(Version)

	app.Flag("debug", "Enable debug logging").BoolVar(&debug)
	app.Flag("logfile", "The file to log to").StringVar(&logfile)
	app.Flag("path", "Path to monitor for metric files").Default(DefaultPath).StringVar(&path)

	exporter := app.Command("export", "Exports the data over HTTP")
	exporter.Flag("port", "The port to listen on ").Default("8080").IntVar(&port)
	exporter.Flag("pid", "Write running PID to a file").StringVar(&pidfile)

	gaugea := app.Command("gauge", "Writes to a gauge metric")
	gaugea.Arg("metric", "The name of the metric to write").Required().StringVar(&metricName)
	gaugea.Arg("value", "The value to write").Required().Float64Var(&value)

	// this command is a typo - guage - its kept here and Hidden() to provide backward compat
	guagea := app.Command("guage", "Writes to a gauge metric").Hidden()
	guagea.Arg("metric", "The name of the metric to write").Required().StringVar(&metricName)
	guagea.Arg("value", "The value to write").Required().Float64Var(&value)

	countera := app.Command("counter", "Increments a counter metric")
	countera.Arg("metric", "The metric name to write").Required().StringVar(&metricName)
	countera.Flag("inc", "How much to increment the counter with").Default("1").FloatVar(&value)

	lista := app.Command("list", "Lists known metrics")
	lista.Arg("filter", "Limit output to metrics containing the filter text").Default("").StringVar(&filter)

	command := kingpin.MustParse(app.Parse(os.Args[1:]))
	log = logrus.NewEntry(logrus.New())

	if logfile != "" {
		file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(fmt.Errorf("Could not set up logging: %s", err))
		}

		log.Logger.SetOutput(file)
	}

	if debug {
		log.Logger.SetLevel(logrus.DebugLevel)
	}

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	go interruptWatcher()

	if pidfile != "" {
		err := ioutil.WriteFile(pidfile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
		if err != nil {
			logrus.Fatalf("Could not write pid file %s: %s", pidfile, err)
		}
	}

	switch command {
	case exporter.FullCommand():
		err = export()
	case gaugea.FullCommand():
		err = gauge()
	case guagea.FullCommand():
		err = gauge()
	case countera.FullCommand():
		err = counter()
	case lista.FullCommand():
		err = list()
	}

	if err != nil {
		logrus.Fatalf("Could not run %s: %s", command, err)
	}
}

func interruptWatcher() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case sig := <-sigs:
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Infof("Shutting down on %s", sig)
				cancel()
			}
		case <-ctx.Done():
			return
		}
	}
}
