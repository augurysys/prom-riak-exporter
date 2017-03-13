package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type gauges map[string]prometheus.Gauge

func (g *gauges) get(name string) prometheus.Gauge {
	if g, ok := (*g)[name]; ok {
		return g
	}

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "riak",
		Name:      name,
		Help:      name,
	})

	(*g)[name] = gauge
	prometheus.MustRegister(gauge)

	return gauge
}

func main() {
	target := flag.String("target", "", "Riak node HTTP url")
	flag.Parse()

	if *target == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	http.Handle("/metrics", prometheus.Handler())

	g := make(gauges)

	go func() {
		for {
			func() {
				resp, err := http.Get(fmt.Sprintf("%s/ping", *target))
				if err != nil {
					g.get("node_up").Set(0)
					return
				}

				defer resp.Body.Close()

				if resp.StatusCode == 200 {
					g.get("node_up").Set(1)
					return
				}

				g.get("node_up").Set(0)
			}()

			func() {
				resp, err := http.Get(fmt.Sprintf("%s/stats", *target))
				if err != nil {
					return
				}

				defer resp.Body.Close()
				if resp.StatusCode != 200 {
					return
				}

				var data map[string]interface{}
				dec := json.NewDecoder(resp.Body)
				if err := dec.Decode(&data); err != nil {
					return
				}

				for k, v := range data {
					if value, ok := v.(float64); ok {
						g.get(k).Set(value)
					}
				}
			}()

			time.Sleep(5 * time.Second)
		}
	}()

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
