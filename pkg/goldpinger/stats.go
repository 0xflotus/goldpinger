// Copyright 2018 Bloomberg Finance L.P.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package goldpinger

import (
	"log"
	"time"

	"github.com/bloomberg/goldpinger/pkg/models"
	"github.com/go-openapi/strfmt"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	goldpingerStatsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goldpinger_stats_total",
			Help: "Statistics of calls made in goldpinger instances",
		},
		[]string{
			"goldpinger_instance",
			"group",
			"action",
		},
	)

	goldpingerNodesHealthGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "goldpinger_nodes_health_total",
			Help: "Number of nodes seen as healthy/unhealthy from this instance's POV",
		},
		[]string{
			"goldpinger_instance",
			"status",
		},
	)

	goldpingerResponseTimePeersHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "goldpinger_peers_response_time_s",
			Help:    "Histogram of response times from other hosts, when making peer calls",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
		},
		[]string{
			"goldpinger_instance",
			"call_type",
			"host_ip",
			"pod_ip",
		},
	)

	goldpingerResponseTimeKubernetesHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "goldpinger_kube_master_response_time_s",
			Help:    "Histogram of response times from kubernetes API server, when listing other instances",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
		},
		[]string{
			"goldpinger_instance",
		},
	)

	goldpingerErrorsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goldpinger_errors_total",
			Help: "Statistics of errors per instance",
		},
		[]string{
			"goldpinger_instance",
			"type",
		},
	)

	groups = map[string]map[string]int64{
		"received": map[string]int64{
			"ping":      0,
			"check":     0,
			"check_all": 0,
		},
		"made": map[string]int64{
			"ping":      0,
			"check":     0,
			"check_all": 0,
		},
	}

	bootTime = time.Now()
)

func init() {
	prometheus.MustRegister(goldpingerStatsCounter)
	prometheus.MustRegister(goldpingerNodesHealthGauge)
	prometheus.MustRegister(goldpingerResponseTimePeersHistogram)
	prometheus.MustRegister(goldpingerResponseTimeKubernetesHistogram)
	prometheus.MustRegister(goldpingerErrorsCounter)
	log.Println("Metrics setup - see /metrics")
}

func GetStats() *models.PingResults {
	var result models.PingResults
	var calls models.CallStats

	calls.Check = groups["received"]["check"]
	calls.CheckAll = groups["received"]["check_all"]
	calls.Ping = groups["received"]["ping"]
	result.BootTime = strfmt.DateTime(bootTime)
	result.Received = &calls
	return &result
}

// counts various calls received and made
func CountCall(group string, call string) {
	groups[group][call]++

	goldpingerStatsCounter.WithLabelValues(
		GoldpingerConfig.Hostname,
		group,
		call,
	).Inc()
}

// counts healthy and unhealthy nodes
func CountHealthyUnhealthyNodes(healthy, unhealthy float64) {
	goldpingerNodesHealthGauge.WithLabelValues(
		GoldpingerConfig.Hostname,
		"healthy",
	).Set(healthy)
	goldpingerNodesHealthGauge.WithLabelValues(
		GoldpingerConfig.Hostname,
		"unhealthy",
	).Set(unhealthy)
}

// counts instances of various errors
func CountError(errorType string) {
	goldpingerErrorsCounter.WithLabelValues(
		GoldpingerConfig.Hostname,
		errorType,
	).Inc()
}

// returns a timer for easy observing of the durations of calls to kubernetes API
func GetLabeledKubernetesCallsTimer() *prometheus.Timer {
	return prometheus.NewTimer(
		goldpingerResponseTimeKubernetesHistogram.WithLabelValues(
			GoldpingerConfig.Hostname,
		),
	)
}

// returns a timer for easy observing of the duration of calls to peers
func GetLabeledPeersCallsTimer(callType, hostIP, podIP string) *prometheus.Timer {
	return prometheus.NewTimer(
		goldpingerResponseTimePeersHistogram.WithLabelValues(
			GoldpingerConfig.Hostname,
			callType,
			hostIP,
			podIP,
		),
	)
}
