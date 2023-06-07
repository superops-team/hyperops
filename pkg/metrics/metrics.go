package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// InfoGauge 当前信息统计
	InfoGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hyperops_info",
			Help: "hyperops info",
		},
		[]string{"version"},
	)

	// ErrorCounter 统计hyperops库本身错误个数统计
	ErrorCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hyperops_error_total",
			Help: "The total number of error message",
		},
		[]string{"reason"},
	)
	// HyperFnCounter 错误函数执行统计，分析失败率
	HyperFnCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hyperops_func_counter",
			Help: "The total number of func call",
		},
		[]string{"func", "status"},
	)

	// HyperFnDurHis 统计每个函数执行时长
	HyperFnDurHis = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hyperops_func_duration_ms",
			Help:    "The total number of func call",
			Buckets: []float64{1, 5, 10, 60, 300, 500, 1000, 3000, 5000},
		},
		[]string{"func", "status"},
	)

	// WorkDuration 执行任务时间
	WorkDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hyperops_work_duration_seconds",
			Help:    "How long in seconds processing an item from workqueue takes.",
			Buckets: []float64{0.5, 1, 5, 10, 60, 300},
		},
		[]string{"name", "status"},
	)
	// WorkCount rate{work_count(job_name="hyperops")}
	WorkCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hyperops_work_count",
			Help: "Count the numbers of the scripts",
		},
		[]string{"name", "status"},
	)
	// HangGouge 挂起的任务统计
	HangGouge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hyperops_hang_count",
			Help: "Count the hanging status task",
		},
		[]string{"name"},
	)
)

func init() {
	InfoGauge.WithLabelValues("v0.1.0").Inc()
}
