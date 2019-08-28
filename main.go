//go_* metrics:35
//promhttp_metric_handler_requests_* metrics: 4

package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	goMetricsNum   = 35
	promMetricsNum = 4
)

var (
	addr      = flag.String("listen-address", ":9200", "The address to listen on for HTTP requests.")
	metricNum = flag.Int("metrics", 0, "The number of metrics")
	labelNum  = flag.Int("labels", 0, "The number of custom labels")
	interval  = flag.Duration("interval", 1*time.Second, "The interval for updating metrics")
	filename  = flag.String("file", "", "The source metrics file.")
)

var (
	counterVec   *prometheus.CounterVec
	counters     []prometheus.Counter
	gaugeVec     *prometheus.GaugeVec
	gauges       []prometheus.Gauge
	summaryVec   *prometheus.SummaryVec
	histogramVec *prometheus.HistogramVec
)

const (
	DefaultNamespace = "monitor"
	DefaultSubsystem = "exporter"
)

func init() {
	// Add Go module build info.
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}

func main() {
	flag.Parse()

	// load metrics from text file
	if *filename != "" {
		fmt.Printf("init metrics from file: %s\n", *filename)
		initMetricsFromFile(*filename)
	}

	if *metricNum > 0 && *labelNum > 0 {
		if *metricNum < 20 {
			*metricNum = 20
			fmt.Printf("minimum metricsNum: %d\n", 20)
		}
		fmt.Printf("generate metrics with metricsNum=%d, labelNum=%d\n", *metricNum, *labelNum)
		// 20 = 4 + 7 + 9
		num := (*metricNum) / 20
		initPromRegister(num, *labelNum)

		go Run(*interval, num, *labelNum)
	}

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(*addr, nil)

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("working....")
	<-sig
	fmt.Println("exit")
}

func initPromRegister(metricNum, labelNum int) {
	counterVec = NewCounterVecWithNum(labelNum)
	counters = NewCounterWithNum(metricNum)
	gaugeVec = NewGaugeVecWithNum(labelNum)
	gauges = NewGaugeWithNum(metricNum)
	summaryVec = NewSummaryVecWithNum(labelNum)
	histogramVec = NewHistogramVecWithNum(labelNum)

	collectors := make([]prometheus.Collector, 0)
	collectors = append(collectors, counterVec, gaugeVec, summaryVec, histogramVec)
	for i := 0; i < len(counters); i++ {
		collectors = append(collectors, counters[i])
	}
	for i := 0; i < len(gauges); i++ {
		collectors = append(collectors, gauges[i])
	}

	prometheus.MustRegister(collectors...)
}

func Run(interval time.Duration, metricNum, labelNum int) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			run(metricNum, labelNum)
		}
	}
}

func run(metricNum, labelNum int) {
	randValue := rand.Float64() * 10

	for i := 0; i < metricNum; i++ {
		labelValue := newLableValues(labelNum, i)
		counterVec.WithLabelValues(labelValue...).Inc()
		gaugeVec.WithLabelValues(labelValue...).Set(randValue)
		summaryVec.WithLabelValues(labelValue...).Observe(randValue)
		histogramVec.WithLabelValues(labelValue...).Observe(randValue)
	}

	for i := range counters {
		counters[i].Inc()
	}
	for i := range gauges {
		gauges[i].Set(randValue)
	}
}

func newLabelNames(labelNum int) []string {
	labelNames := NewLabelNamesWithNum(labelNum)
	labelNames = append(labelNames, "index")
	return labelNames
}

func NewLabelNamesWithNum(labelNum int) []string {
	labelNames := make([]string, labelNum)
	for i := 0; i < labelNum; i++ {
		labelNames[i] = fmt.Sprintf("key%d", i)
	}
	return labelNames
}

func newLableValues(labelNum int, index int) []string {
	labelValues := NewLabelValuesWithNum(labelNum)
	labelValues = append(labelValues, strconv.FormatInt(int64(index), 10))
	return labelValues
}

func NewLabelValuesWithNum(labelNum int) []string {
	labelValues := make([]string, labelNum)
	for i := 0; i < labelNum; i++ {
		labelValues[i] = fmt.Sprintf("value%d", i)
	}
	return labelValues
}

func NewCounterVecWithNum(labelNum int) *prometheus.CounterVec {
	labelNames := newLabelNames(labelNum)
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: DefaultNamespace,
		Subsystem: DefaultSubsystem,
		Name:      "request_total",
		Help:      "The total number of requests",
	}, labelNames)
	return counter
}

func NewCounterWithNum(metricNum int) []prometheus.Counter {
	counters := make([]prometheus.Counter, 0, metricNum)
	for i := 0; i < metricNum; i++ {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   DefaultNamespace,
			Subsystem:   DefaultSubsystem,
			Name:        fmt.Sprintf("response_%d_total", i),
			Help:        "The total number of responses",
			ConstLabels: map[string]string{"repo": "custom-exporter"},
		})
		counters = append(counters, counter)
	}
	return counters
}

func NewGaugeVecWithNum(labelNum int) *prometheus.GaugeVec {
	labelNames := newLabelNames(labelNum)
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: DefaultNamespace,
		Subsystem: DefaultSubsystem,
		Name:      "request_duration_seconds",
		Help:      "The process duration of the request",
	}, labelNames)
	return gauge
}

func NewGaugeWithNum(metricNum int) []prometheus.Gauge {
	gauges := make([]prometheus.Gauge, 0, metricNum)
	for i := 0; i < metricNum; i++ {
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace:   DefaultNamespace,
			Subsystem:   DefaultSubsystem,
			Name:        fmt.Sprintf("memory_%d_usage_bytes", i),
			Help:        "The usage bytes of the memory",
			ConstLabels: map[string]string{"repo": "custom-exporter"},
		})
		gauges = append(gauges, gauge)
	}
	return gauges
}

func NewSummaryVecWithNum(labelNum int) *prometheus.SummaryVec {
	labelNames := newLabelNames(labelNum)
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  DefaultNamespace,
		Subsystem:  DefaultSubsystem,
		Name:       "rpc_duration_seconds",
		Help:       "RPC latency distributions.",
		Objectives: map[float64]float64{0.1: 0.5, 0.5: 0.05, 0.9: 0.01, 0.99: 0.005, 1: 0.001},
	}, labelNames)
	return summary
}

func NewHistogramVecWithNum(labelNum int) *prometheus.HistogramVec {
	labelNames := newLabelNames(labelNum)
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: DefaultNamespace,
		Subsystem: DefaultSubsystem,
		Name:      "rpc_durations_histogram_seconds",
		Help:      "RPC latency distributions.",
		Buckets:   []float64{0.1, 1, 3, 5, 7, 10},
	}, labelNames)
	return histogram
}
