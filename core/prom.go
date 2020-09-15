package core

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
)

func updateGuage(corporation *AgentableCorporation,
	TOAvailable *prometheus.GaugeVec,
	TOUsed *prometheus.GaugeVec,
	TOTotal *prometheus.GaugeVec) {
	updateElement := func(guage *prometheus.GaugeVec, value int) {
		guage.With(prometheus.Labels{
			"name":         corporation.Name,
			"corpSize":     corporation.CorporationSize,
			"buisnessType": corporation.BusinessType,
		}).Set(float64(value))
	}
	updateElement(TOAvailable, corporation.HyunyukBaejung)
	updateElement(TOUsed, corporation.HyunyukPyunip)
	updateElement(TOTotal, corporation.HyunukBokmu)
}

func CrawlAndUpdateGuage(
	TOAvailable *prometheus.GaugeVec,
	TOUsed *prometheus.GaugeVec,
	TOTotal *prometheus.GaugeVec) {
	crawlResult := crawlAllCorporations()
	fmt.Println(crawlResult)
	for _, corp := range crawlResult {
		updateGuage(&corp, TOAvailable, TOUsed, TOTotal)
	}
}
