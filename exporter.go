package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ocowchun/taipower_exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	wait = time.Second * 15
)

func main() {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics and web interface.",
		).Default(":8080").String()
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
	)
	kingpin.Parse()

	registry := prometheus.NewRegistry()

	registry.MustRegister(collector.New())
	http.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Taipower Exporter</title></head>
			<body>
			<h1>Taipower Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	server := http.Server{}
	listener, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		log.Fatalf("Could not create listener: %v", err)
		os.Exit(1)
	}
	log.Printf("Listening on %s", *listenAddress)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Signal received: %v. Exiting...", <-signalChan)
		ctx, cancel := context.WithTimeout(context.Background(), wait)
		defer cancel()
		log.Println("shutting down")
		server.Shutdown(ctx)
	}()

	log.Fatal(server.Serve(listener))
}
