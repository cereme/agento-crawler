package core

import (
	"github.com/prometheus/client_golang/prometheus"
)

type GaugeSet struct {
	TOAvailable *prometheus.GaugeVec
	TOUsed      *prometheus.GaugeVec
	TOTotal     *prometheus.GaugeVec
}

func updateGuage(corporation *AgentableCorporation, gauges *GaugeSet) {
	updateElement := func(guage *prometheus.GaugeVec, value int) {
		guage.With(prometheus.Labels{
			"name":         corporation.Name,
			"corpSize":     corporation.CorporationSize,
			"buisnessType": corporation.BusinessType,
		}).Set(float64(value))
	}
	updateElement(gauges.TOAvailable, corporation.HyunyukBaejung)
	updateElement(gauges.TOUsed, corporation.HyunyukPyunip)
	updateElement(gauges.TOTotal, corporation.HyunukBokmu)
}

func CrawlAndUpdateGuage(gauges *GaugeSet) {
	crawlResult := crawlAllCorporations()
	for _, corp := range crawlResult {
		updateGuage(&corp, gauges)
	}
}
