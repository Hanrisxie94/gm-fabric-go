package prometheus

// Copyright 2017 Decipher Technology Studios LLC
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

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/deciphernow/gm-fabric-go/metrics/flatjson"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// PromReporter implements the Reporter interface
type PromReporter struct {
	PrometheusURI string
}

// Report implements the Reporter interface
func (rpt *PromReporter) Report(jWriter *flatjson.Writer) error {
	var err error
	reportMap := make(reportMapType)

	if err = accumulateMetrics(rpt.PrometheusURI, reportMap); err != nil {
		return errors.Wrap(err, "accumulateMetrics")
	}

	if err = reportMetrics(jWriter, reportMap); err != nil {
		return errors.Wrap(err, "reportMetrics")
	}

	return nil
}

// httpStatus is a 3 character string of the form 100-599
type httpStatus string

type reportEntry struct {
	// requests       uint64
	statusCount map[httpStatus]uint64
	//	latencyMsCount uint64
	//	latencyMsMax   float64
	//	latencyMsMin   float64
	latencyMsSum   float64
	latencyMsP50   float64
	latencyMsP90   float64
	latencyMsP95   float64
	latencyMsP99   float64
	latencyMsP9990 float64
	latencyMsP9999 float64
	inThroughput   uint64
	outThroughput  uint64
}

func reportAll(key model.LabelValue) bool {
	return key == model.LabelValue(AllMetricsKey)
}

type reportKey struct {
	key    model.LabelValue
	method model.LabelValue
}

type reportMapType map[reportKey]reportEntry

func computeReportKey(m *model.Sample) reportKey {
	if reportAll(m.Metric["key"]) {
		return reportKey{
			key: m.Metric["key"],
		}
	}
	return reportKey{
		key:    m.Metric["key"],
		method: m.Metric["method"],
	}
}

type sampleFuncType func(*model.Sample, *reportEntry) error

func accumulateMetrics(
	prometheusURI string,
	reportMap reportMapType,
) error {
	var err error
	type accumulateRequestType struct {
		query      string
		sampleFunc sampleFuncType
	}
	accumulateRequests := []accumulateRequestType{
		accumulateRequestType{
			query:      "http_request_duration_seconds",
			sampleFunc: durationSampleFunc,
		},
		accumulateRequestType{
			query:      "http_request_duration_seconds_count",
			sampleFunc: countSampleFunc,
		},
		accumulateRequestType{
			query:      "http_request_duration_seconds_sum",
			sampleFunc: sumSampleFunc,
		},
		accumulateRequestType{
			query:      "http_request_size_bytes",
			sampleFunc: inSampleFunc,
		},
		accumulateRequestType{
			query:      "http_response_size_bytes",
			sampleFunc: outSampleFunc,
		},
	}

	client, err := api.NewClient(
		api.Config{Address: prometheusURI},
	)
	if err != nil {
		return errors.Wrap(err, "api.NewClient")
	}

	promAPI := v1.NewAPI(client)

	timeStamp := time.Now()

	for i, accumulateRequest := range accumulateRequests {
		err = accumulateReport(
			promAPI,
			timeStamp,
			accumulateRequest.query,
			reportMap,
			accumulateRequest.sampleFunc,
		)
		if err != nil {
			return errors.Wrapf(err, "#%d: accumulateReport: %s", i+1, accumulateRequest.query)
		}
	}

	return nil
}

func durationSampleFunc(m *model.Sample, e *reportEntry) error {
	ms := m.Value * 1000
	switch m.Metric["quantile"] {
	case "0.5":
		e.latencyMsP50 = float64(ms)
	case "0.9":
		e.latencyMsP90 = float64(ms)
	case "0.95":
		e.latencyMsP95 = float64(ms)
	case "0.99":
		e.latencyMsP99 = float64(ms)
	case "0.999":
		e.latencyMsP9990 = float64(ms)
	case "0.9999":
		e.latencyMsP9999 = float64(ms)
	default:
		return errors.Errorf("unknown quantile '%s'", m.Metric["quantile"])
	}
	return nil
}

func countSampleFunc(m *model.Sample, e *reportEntry) error {
	rawStatus := string(m.Metric["status"])
	numericStatus, err := strconv.Atoi(rawStatus)
	if err != nil {
		return errors.Wrapf(err, "strconv.Atoi(%s)", rawStatus)
	}
	if numericStatus < 100 || numericStatus > 599 {
		return errors.Errorf("Invalid HTTP status %d", numericStatus)
	}
	cleanStatus := httpStatus(fmt.Sprintf("%03d", numericStatus))

	e.statusCount[cleanStatus] = uint64(m.Value)

	return nil
}

func sumSampleFunc(m *model.Sample, e *reportEntry) error {
	e.latencyMsSum = float64(m.Value) * 1000
	return nil
}

func inSampleFunc(m *model.Sample, e *reportEntry) error {
	e.inThroughput = uint64(m.Value)
	return nil
}

func outSampleFunc(m *model.Sample, e *reportEntry) error {
	e.outThroughput = uint64(m.Value)
	return nil
}

func accumulateReport(
	promAPI v1.API,
	timeStamp time.Time,
	query string,
	reportMap reportMapType,
	sampleFunc sampleFuncType,
) error {
	var err error

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()

	value, err := promAPI.Query(ctx, query, timeStamp)
	if err != nil {
		return errors.Wrapf(err, "promAPI.Query: %s", query)
	}

	vectorValue, ok := value.(model.Vector)
	if !ok {
		return errors.Errorf("%s: unable to cast %T to vector", query, value)
	}

VECTOR_LOOP:
	for _, value := range vectorValue {
		// expedient to weed out Prometheus metrics with the same name
		if value.Metric["key"] == "" {
			continue VECTOR_LOOP
		}
		rk := computeReportKey(value)
		rm, ok := reportMap[rk]
		if !ok {
			rm.statusCount = make(map[httpStatus]uint64)
		}
		if err = sampleFunc(value, &rm); err != nil {
			return errors.Wrap(err, "sampleFunc")
		}
		reportMap[rk] = rm
	}

	return nil
}

func reportMetrics(
	jWriter *flatjson.Writer,
	reportMap reportMapType,
) error {
	var err error

	type routeLineType struct {
		label string
		value interface{}
	}

	for mapKey, entry := range reportMap {
		var route string
		if reportAll(mapKey.key) {
			route = fmt.Sprintf("%s/", mapKey.key)
		} else {
			route = fmt.Sprintf("route%s/%s/", mapKey.key, mapKey.method)
		}

		var requestCount uint64
		summaryCount := make(map[httpStatus]uint64)
		var statusLines []routeLineType

		// we assume status is a 3 character string, with char 0 = 1..5
		for status, statusCount := range entry.statusCount {
			summaryStatus := status[:1] + "XX"
			summaryCount[summaryStatus] += statusCount

			requestCount += statusCount

			// "route/acme/services/catalog/GET/status/200": 122,
			statusLine := routeLineType{
				label: fmt.Sprintf("status/%s", status),
				value: statusCount,
			}

			statusLines = append(statusLines, statusLine)
		}

		for summaryStatus, summaryStatusCount := range summaryCount {
			// "route/acme/services/catalog/GET/status/200": 122,
			statusLine := routeLineType{
				label: fmt.Sprintf("status/%s", summaryStatus),
				value: summaryStatusCount,
			}

			statusLines = append(statusLines, statusLine)
		}

		routeLines := []routeLineType{
			// "route/acme/services/catalog/GET/requests": 122,
			routeLineType{label: "requests", value: requestCount},
		}

		routeLines = append(routeLines, statusLines...)

		var latencyMsAvg float64
		if requestCount > 0 {
			latencyMsAvg = entry.latencyMsSum / float64(requestCount)
		}

		routeLines = append(routeLines, []routeLineType{
			// "route/acme/services/catalog/GET/routes": "",
			routeLineType{label: "routes", value: ""},
			// "route/acme/services/catalog/GET/latency_ms.avg": 1206.598361,
			routeLineType{label: "latency_ms.avg", value: latencyMsAvg},
			// "route/acme/services/catalog/GET/latency_ms.count": 122,
			routeLineType{label: "latency_ms.count", value: requestCount},
			// "route/acme/services/catalog/GET/latency_ms.max": 1968,
			routeLineType{label: "latency_ms.max", value: 0},
			// "route/acme/services/catalog/GET/latency_ms.min": 513,
			routeLineType{label: "latency_ms.min", value: 0},
			// "route/acme/services/catalog/GET/latency_ms.sum": 147205,
			routeLineType{label: "latency_ms.sum", value: uint64(entry.latencyMsSum)},
			// "route/acme/services/catalog/GET/latency_ms.p50": 1172,
			routeLineType{label: "latency_ms.p50", value: uint64(entry.latencyMsP50)},
			// "route/acme/services/catalog/GET/latency_ms.p90": 1757,
			routeLineType{label: "latency_ms.p90", value: uint64(entry.latencyMsP90)},
			// "route/acme/services/catalog/GET/latency_ms.p95": 1825,
			routeLineType{label: "latency_ms.p95", value: uint64(entry.latencyMsP95)},
			// "route/acme/services/catalog/GET/latency_ms.p99": 1923,
			routeLineType{label: "latency_ms.p99", value: uint64(entry.latencyMsP99)},
			// "route/acme/services/catalog/GET/latency_ms.p9990": 1968,
			routeLineType{label: "latency_ms.p9990", value: uint64(entry.latencyMsP9990)},
			// "route/acme/services/catalog/GET/latency_ms.p9999": 1968,
			routeLineType{label: "latency_ms.p9999", value: uint64(entry.latencyMsP9999)},
			// "route/acme/services/catalog/GET/in_throughput": 0,
			routeLineType{label: "in_throughput", value: entry.inThroughput},
			// "route/acme/services/catalog/GET/out_throughput": 71287,
			routeLineType{label: "out_throughput", value: entry.outThroughput},
		}...)
		for _, routeLine := range routeLines {
			err = jWriter.Write(
				fmt.Sprintf("%s%s", route, routeLine.label),
				routeLine.value,
			)
			if err != nil {
				return errors.Wrap(err, "flatjson.New")
			}
		}

		/*
			"route/acme/services/catalog/GET/errors.count": 0,
		*/
	}

	return nil
}
