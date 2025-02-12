/*
 * Copyright The Dragonfly Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"net/http"

	"github.com/dragonflyoss/Dragonfly/pkg/metricsutils"
	"github.com/dragonflyoss/Dragonfly/supernode/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// metrics defines some prometheus metrics for monitoring supernode
type metrics struct {
	requestCounter  *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestSize     *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
}

func newMetrics(register prometheus.Registerer) *metrics {
	return &metrics{
		requestCounter: metricsutils.NewCounter(config.SubsystemSupernode, "http_requests_total",
			"Counter of HTTP requests.", []string{"code", "handler"}, register,
		),
		requestDuration: metricsutils.NewHistogram(config.SubsystemSupernode, "http_request_duration_seconds",
			"Histogram of latencies for HTTP requests.", []string{"handler"},
			[]float64{.1, .2, .4, 1, 3, 8, 20, 60, 120}, register,
		),
		requestSize: metricsutils.NewHistogram(config.SubsystemSupernode, "http_request_size_bytes",
			"Histogram of request size for HTTP requests.", []string{"handler"},
			prometheus.ExponentialBuckets(100, 10, 8), register,
		),
		responseSize: metricsutils.NewHistogram(config.SubsystemSupernode, "http_response_size_bytes",
			"Histogram of response size for HTTP requests.", []string{"handler"},
			prometheus.ExponentialBuckets(100, 10, 8), register,
		),
	}
}

// instrumentHandler will update metrics for every http request
func (m *metrics) instrumentHandler(handlerName string, handler http.HandlerFunc) http.HandlerFunc {
	return promhttp.InstrumentHandlerDuration(
		m.requestDuration.MustCurryWith(prometheus.Labels{"handler": handlerName}),
		promhttp.InstrumentHandlerCounter(
			m.requestCounter.MustCurryWith(prometheus.Labels{"handler": handlerName}),
			promhttp.InstrumentHandlerRequestSize(
				m.requestSize.MustCurryWith(prometheus.Labels{"handler": handlerName}),
				promhttp.InstrumentHandlerResponseSize(
					m.responseSize.MustCurryWith(prometheus.Labels{"handler": handlerName}),
					handler,
				),
			),
		),
	)
}
