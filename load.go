package main

import (
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func initMetricsFromFile(filename string) error {
	metricFamilies, err := getMetricFamiliesFromFile(filename)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	dumpMetricsFromFamilies(metricFamilies)
	return nil
}

func getMetricFamiliesFromFile(filename string) (map[string]*dto.MetricFamily, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %s", filename, err)
	}
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(f)
	if err != nil {
		return nil, fmt.Errorf("failed to parse text metrics: %s", err)
	}
	return metricFamilies, nil
}

func dumpMetricsFromFamilies(metricFamilies map[string]*dto.MetricFamily) {
	for name, metricFamily := range metricFamilies {
		labelKeys, err := getMetricFamilyLabelKeys(metricFamily)
		if err != nil {
			fmt.Printf("failed to get metric family labels: %s\n", err)
			continue
		}
		switch tye := metricFamily.GetType(); tye {
		case dto.MetricType_COUNTER:
			counter := prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: name,
				Help: metricFamily.GetHelp(),
			}, labelKeys)
			for _, metric := range metricFamily.Metric {
				labels := labelPairsToMap(metric.GetLabel())
				counter.With(labels).Write(metric)
			}
		case dto.MetricType_GAUGE:
			gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: name,
				Help: metricFamily.GetHelp(),
			}, labelKeys)
			if err := prometheus.Register(gauge); err != nil {
				fmt.Printf("failed to register %s: [%s %v]\n", tye, name, labelKeys)
				continue
			}
			for _, metric := range metricFamily.Metric {
				labels := labelPairsToMap(metric.GetLabel())
				if len(labels) != len(labelKeys) {
					fmt.Printf("illegal labels %+v\n", labels)
					continue
				}
				gauge.With(labels).Set(*metric.GetGauge().Value)
			}
		case dto.MetricType_HISTOGRAM:
			prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: name,
				Help: metricFamily.GetHelp(),
			}, labelKeys)
		case dto.MetricType_SUMMARY:
			prometheus.NewSummaryVec(prometheus.SummaryOpts{
				Name: name,
				Help: metricFamily.GetHelp(),
			}, labelKeys)
		default:
		}
	}
}

func getMetricFamilyLabelKeys(family *dto.MetricFamily) ([]string, error) {
	labelKeyMap := make(map[string]struct{})
	for _, metric := range family.Metric {
		for _, pair := range metric.GetLabel() {
			labelKeyMap[pair.GetName()] = struct{}{}
		}
	}
	labelKeys, err := GetMapStrKeys(labelKeyMap)
	if err != nil {
		return nil, err
	}
	return labelKeys, nil
}

func labelPairsToMap(pairs []*dto.LabelPair) map[string]string {
	labels := make(map[string]string)
	for _, pair := range pairs {
		labels[*pair.Name] = *pair.Value
	}
	return labels
}
