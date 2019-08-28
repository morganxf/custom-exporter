package main

import (
	"fmt"
	"testing"
)

func Test_GetMetricsFromFile(t *testing.T) {
	metricFamilies, _ := getMetricFamiliesFromFile("./testdata/metrics.txt")
	for name, family := range metricFamilies {
		fmt.Printf("%+v\n", name)
		fmt.Printf("%+v\n", family)
		for _, metric := range family.Metric {
			fmt.Printf("%+v\n", metric.GetLabel())
			fmt.Printf("%+v\n", metric.GetGauge())
		}
	}
}

func Test_DumpMetrics(t *testing.T) {
	metricsFamilies, _ := getMetricFamiliesFromFile("./testdata/metrics.txt")
	dumpMetricsFromFamilies(metricsFamilies)
}
