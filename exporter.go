package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	target := flag.String("target", "", "Riak node HTTP url")
	flag.Parse()

	if *target == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	http.Handle("/metrics", prometheus.Handler())

	up := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "riak",
		Subsystem: "node",
		Name:      "up",
		Help:      "does Riak respond to HTTP pings",
	})

	prometheus.MustRegister(up)

	go func() {
		for {
			func() {
				resp, err := http.Get(fmt.Sprintf("%s/ping", *target))
				if err != nil {
					up.Set(0)
					return
				}

				defer resp.Body.Close()

				if resp.StatusCode == 200 {
					up.Set(1)
					return
				}

				up.Set(0)
			}()

			time.Sleep(5 * time.Second)
		}
	}()

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
