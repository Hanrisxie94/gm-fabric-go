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

package sinkobserver

import (
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	gometrics "github.com/armon/go-metrics"

	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/memvalues"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
)

var (
	validPrometheusRegex   *regexp.Regexp
	invalidPrometheusRegex *regexp.Regexp
)

func init() {
	validPrometheusRegex = regexp.MustCompile("^[a-zA-Z_:]([a-zA-Z0-9_:])*$")
	invalidPrometheusRegex = regexp.MustCompile("[^a-zA-Z0-9_:]")
}

type activeEntry struct {
	stats  grpcobserver.APIStats
	tagMap map[string]string
}

type sinkObs struct {
	sync.Mutex
	active map[string]activeEntry
	sink   gometrics.MetricSink
}

// New return an observer that feeds the go-metrics sink
func New(
	sink gometrics.MetricSink,
	reportInterval time.Duration,
) subject.Observer {
	obs := sinkObs{
		sink:   sink,
		active: make(map[string]activeEntry),
	}
	go obs.reportMemory(reportInterval)
	return &obs
}

// Observe implements the Observer pattern
func (so *sinkObs) Observe(event subject.MetricsEvent) {
	so.Lock()
	defer so.Unlock()

	entry := so.active[event.RequestID]

	// TODO: we are using the APIStats object here because that's
	// what the system was originally designed to collect. We should
	// probably either move it to it's own package, or use different handlers
	entryStats, end := grpcobserver.Accumulate(entry.stats, event)
	entry.stats = entryStats
	if entry.tagMap == nil {
		entry.tagMap = make(map[string]string)
	}
	updateTagMap(entry.tagMap, event)

	if end {
		key := []string{
			entry.tagMap["service"],
			entry.tagMap["host"],
			fixEntryKey(entry.stats.Key),
		}
		elapsed := entry.stats.EndTime.Sub(entry.stats.BeginTime)
		so.sink.IncrCounter(
			append(key, "in_throughput"),
			float32(entry.stats.InWireLength),
		)
		so.sink.IncrCounter(
			append(key, "out_throughput"),
			float32(entry.stats.OutWireLength),
		)
		if entry.stats.Err != nil {
			so.sink.IncrCounter(append(key, "errors"), 1)
		}
		so.sink.AddSample(
			append(key, "latency"),
			duration2ms(elapsed),
		)
		delete(so.active, event.RequestID)
	} else {
		so.active[event.RequestID] = entry
	}
}

func (so *sinkObs) reportMemory(reportInterval time.Duration) {
	tickChan := time.Tick(reportInterval)
MEM_LOOP:
	for {
		<-tickChan
		memValues, err := memvalues.GetMemValues()
		if err != nil {
			log.Printf("ERROR: memvalues.GetMemValues(): %s", err)
			continue MEM_LOOP
		}
		so.Lock()
		so.sink.AddSample(
			[]string{"memory", "system", "available"},
			float32(memValues.SystemMemoryAvailable),
		)
		so.sink.AddSample(
			[]string{"memory", "system", "used"},
			float32(memValues.SystemMemoryUsed),
		)
		so.sink.AddSample(
			[]string{"memory", "system", "used-percent"},
			float32(memValues.SystemMemoryUsedPercent),
		)
		so.sink.AddSample(
			[]string{"memory", "process", "used"},
			float32(memValues.ProcessMemoryUsed),
		)
		so.Unlock()
	}
}

func duration2ms(d time.Duration) float32 {
	const nsPerMs = 1000000

	return float32(d.Nanoseconds() / nsPerMs)
}

func updateTagMap(tagMap map[string]string, event subject.MetricsEvent) {
	for _, tag := range event.Tags {
		name, value := subject.SplitTag(tag)
		if value != "" {
			_, ok := tagMap[name]
			if !ok {
				tagMap[name] = value
			}
		}
	}
}

// fixEntryKey cleans up the key for Prometheus
//
// Prometheus metric names have to adhere to this regex:
// [a-zA-Z_:]([a-zA-Z0-9_:])*
//
// we have a (dashboard)key of the form
// route/repos/deciphernow/bouncycastle-maven-plugin/issues/GET/latency_ms.avg
//
// we want to produce a (temporary expedient) key acceptable to Prometheus
// route:repos_deciphernow_bouncycastle-maven-plugin_issues:GET:latency_ms_avg
//
// for grpc we have a dashboard key of the form
// function/HelloStream/errors.count
//
// we want to produce a (temporary expedient) key acceptable to Prometheus
// function:HelloStream:errors_count
func fixEntryKey(rawKey string) string {
	// if Prometheus will accept this key as-is, don't change it
	if matched := validPrometheusRegex.MatchString(rawKey); matched {
		return rawKey
	}

	splitKey := strings.Split(rawKey, "/")

	// if this key doesn't have slashes, and Prometheus doesn't like it,
	// we can't fix it.
	if len(splitKey) == 1 {
		log.Printf("ERROR: unsplittable invalid key for Prometheus: %s", rawKey)
		return fixInvalidKey(rawKey)
	}

	return fixSplitKey(rawKey, splitKey)
}

func fixSplitKey(rawKey string, splitKey []string) string {

	var fixedKey string

	switch splitKey[0] {
	case "function":
		// "function/HelloStream/errors.count
		if len(splitKey) != 2 {
			log.Printf("ERROR: invalid function key for Prometheus: %s", rawKey)
			return fixInvalidKey(rawKey)
		}
		// assume we have 'function', function-name
		functionName := splitKey[1]
		fixedKey = strings.Join(
			[]string{
				"function",
				invalidPrometheusRegex.ReplaceAllLiteralString(functionName, "_"),
			},
			":",
		)
	case "route":
		// route/repos/deciphernow/bouncycastle-maven-plugin/issues/GET/latency_ms.avg
		if len(splitKey) < 3 {
			log.Printf("ERROR: invalid route key for Prometheus: %s", rawKey)
			return fixInvalidKey(rawKey)
		}
		// assume we have 'route', uri[0]...uri[n], method, metric
		uri := strings.Join(splitKey[1:len(splitKey)-1], "_")
		method := splitKey[len(splitKey)-1]
		fixedKey = strings.Join(
			[]string{
				"route",
				invalidPrometheusRegex.ReplaceAllLiteralString(uri, "_"),
				invalidPrometheusRegex.ReplaceAllLiteralString(method, "_"),
			},
			":",
		)
	default:
		log.Printf("ERROR: unknown key for Prometheus: %s", rawKey)
		fixedKey = fixInvalidKey(rawKey)
	}

	if matched := validPrometheusRegex.MatchString(fixedKey); !matched {
		log.Printf("ERROR: unfixable key for Prometheus: %s, %s", rawKey, fixedKey)
		fixedKey = fixInvalidKey(rawKey)
	}

	return fixedKey
}

// fixInvalidKey brute force replaces all invalid characters in a key
// so we can at least see the bad key
func fixInvalidKey(rawKey string) string {
	s := invalidPrometheusRegex.ReplaceAllLiteralString(rawKey, "_")
	// the first character can't be numeric, so just brute force padd it
	return "x" + s
}
