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

/*
cloudwatch provides a metric publisher. To publish metrics, you'll use some
pattern similar to:

	// Set up your cloudwatch client
	var client *cloudwatch.CloudWatch

	// Assign some service name here
	serviceName string

	// Assign a cloudwatch namespace here
	namespace string

	// A callback used for collecting your stats.
	var snapshot := func() *grpcobserver.APIEndpointStats {
		var stats *grpcobserver.APIEndpointStats
		// ... this is where you'd collect your API stats ...
		return stats
	}

	// Create your MetricPublisher
	publisher, err := NewMetricPublisher(client, snapshot, serviceName, namespace)

	publisher.PublishMetrics()
*/
package cloudwatch

import (
	"errors"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"

	"github.com/shirou/gopsutil/cpu"
	memutil "github.com/shirou/gopsutil/mem"
)

type MetricPublisher struct {
	client      *cloudwatch.CloudWatch
	serviceName string
	namespace   string
	snapshot    func() *grpcobserver.APIEndpointStats
}

// Create a MetricPublisher from a given CloudWatch client, metrics,
// service name and namespace.
func NewMetricPublisher(client *cloudwatch.CloudWatch, metricsSnapshot func() *grpcobserver.APIEndpointStats, serviceName string, namespace string) (*MetricPublisher, error) {
	if client == nil {
		return nil, errors.New("CloudWatch client is nil")
	}

	pub := MetricPublisher{
		client:      client,
		snapshot:    metricsSnapshot,
		serviceName: serviceName,
		namespace:   namespace,
	}

	return &pub, nil
}

// Publish metrics.
func (self *MetricPublisher) PublishMetrics() error {
	var cwSession *cloudwatch.CloudWatch
	vmem, _ := memutil.VirtualMemory()
	var metricDatum []*cloudwatch.MetricDatum
	now := aws.Time(time.Now())

	var dims []*cloudwatch.Dimension
	dims = append(dims, &cloudwatch.Dimension{Name: aws.String("Service Name"), Value: aws.String(self.serviceName)})

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	usedPercent := 100.0 * (float64(mem.Alloc) / float64(vmem.Total))
	metricDatum = append(metricDatum,
		&cloudwatch.MetricDatum{
			MetricName: aws.String("process/memory/percent"),
			Dimensions: dims,
			Timestamp:  now,
			Unit:       aws.String("None"),
			Value:      &usedPercent,
		},
	)

	usedMemKB := float64(mem.Alloc / 1024)
	metricDatum = append(metricDatum,
		&cloudwatch.MetricDatum{
			MetricName: aws.String("process/memory/kb"),
			Dimensions: dims,
			Timestamp:  now,
			Unit:       aws.String("None"),
			Value:      &usedMemKB,
		},
	)

	// cpu.Percent() returns a number in the range 0..100.0
	var cpuPercent float64
	percents, err := cpu.Percent(0, false)
	if err == nil && len(percents) == 1 {
		cpuPercent = percents[0]
	}
	metricDatum = append(metricDatum,
		&cloudwatch.MetricDatum{
			MetricName: aws.String("process/cpu/percent"),
			Dimensions: dims,
			Timestamp:  now,
			Unit:       aws.String("None"),
			Value:      &cpuPercent,
		},
	)

	if stats := self.snapshot(); stats != nil {
		var p90 float64
		p90 = float64(stats.P90)

		metricDatum = append(metricDatum,
			&cloudwatch.MetricDatum{
				MetricName: aws.String("srv/request_latency_ms.p90"),
				Dimensions: dims,
				Timestamp:  now,
				Unit:       aws.String("None"),
				Value:      &p90,
			},
		)
	}

	params := &cloudwatch.PutMetricDataInput{
		Namespace:  &self.namespace,
		MetricData: metricDatum,
	}

	_, err = cwSession.PutMetricData(params)
	return err
}
