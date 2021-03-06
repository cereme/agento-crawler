package main

import (
	"agento-crawler/core"
	"github.com/mileusna/crontab"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var (
	TOAvailable = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "TOAvailable",
	}, []string{"name", "corpSize", "buisnessType"})
	TOUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "TOUsed",
	}, []string{"name", "corpSize", "buisnessType"})
	TOTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "TOTotal",
	}, []string{"name", "corpSize", "buisnessType"})

	LastCrawledAt = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "LastCrawledAt",
	})
	LastCrawlTimeSpent = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "LastCrawlTimeSpent",
	})
)

func init() {
	prometheus.MustRegister(TOAvailable)
	prometheus.MustRegister(TOUsed)
	prometheus.MustRegister(TOTotal)

	prometheus.MustRegister(LastCrawledAt)
	prometheus.MustRegister(LastCrawlTimeSpent)
}

func main() {
	ctab := crontab.New()

	gauges := new(core.GaugeSet)
	gauges.TOAvailable = TOAvailable
	gauges.TOUsed = TOUsed
	gauges.TOTotal = TOTotal
	gauges.LastCrawlTimeSpent = &LastCrawlTimeSpent
	gauges.LastCrawledAt = &LastCrawledAt

	core.CrawlAndUpdateGuage(gauges)

	_ = ctab.AddJob("0 0 * * *", func() { core.CrawlAndUpdateGuage(gauges) })
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
