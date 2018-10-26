package cloudobserver

import (
	"fmt"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/memvalues"
	"github.com/shirou/gopsutil/cpu"
)

type datumFuncType func(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum

var datumFuncMap = map[string]datumFuncType{
	"latency_ms.avg":             latencyMSAvg,
	"latency_ms.count":           latencyMSCount,
	"latency_ms.max":             latencyMSMax,
	"latency_ms.min":             latencyMSMin,
	"latency_ms.sum":             latencyMSSum,
	"latency_ms.p50":             latencyMSP50,
	"latency_ms.p90":             latencyMSP90,
	"latency_ms.p95":             latencyMSP95,
	"latency_ms.p99":             latencyMSP99,
	"latency_ms.p9990":           latencyMSP9990,
	"latency_ms.p9999":           latencyMSP9999,
	"errors.count":               errorsCount,
	"in_throughput":              inThroughput,
	"out_throughput":             outThroughput,
	"system/cpu.pct":             systemCPUPct,
	"system/cpu_cores":           systemCPUCores,
	"system/memory/available":    systemMemoryAvailable,
	"system/memory/used":         systemMemoryUsed,
	"system/memory/used_percent": systemMemoryUsedPercent,
	"process/memory/used":        processMemoryUsed,
}

func latencyMSAvg(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "latency_ms.avg")),
		Unit:       aws.String("Milliseconds"),
		Value:      aws.Float64(stats[key].Avg),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func latencyMSCount(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "latency_ms.count")),
		Unit:       aws.String("Count"),
		Value:      aws.Float64(float64(stats[key].Count)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func latencyMSMax(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "latency_ms.max")),
		Unit:       aws.String("Milliseconds"),
		Value:      aws.Float64(float64(stats[key].Max)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func latencyMSMin(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "latency_ms.min")),
		Unit:       aws.String("Milliseconds"),
		Value:      aws.Float64(float64(stats[key].Min)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func latencyMSSum(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "latency_ms.sum")),
		Unit:       aws.String("Milliseconds"),
		Value:      aws.Float64(float64(stats[key].Sum)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func latencyMSP50(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "latency_ms.p50")),
		Unit:       aws.String("Count"),
		Value:      aws.Float64(float64(stats[key].P50)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func latencyMSP90(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "latency_ms.p90")),
		Unit:       aws.String("Count"),
		Value:      aws.Float64(float64(stats[key].P90)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func latencyMSP95(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "latency_ms.p95")),
		Unit:       aws.String("Count"),
		Value:      aws.Float64(float64(stats[key].P95)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func latencyMSP99(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "latency_ms.p99")),
		Unit:       aws.String("Count"),
		Value:      aws.Float64(float64(stats[key].P99)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func latencyMSP9990(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "latency_ms.p9990")),
		Unit:       aws.String("Count"),
		Value:      aws.Float64(float64(stats[key].P9990)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func latencyMSP9999(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "latency_ms.p9999")),
		Unit:       aws.String("Count"),
		Value:      aws.Float64(float64(stats[key].P9999)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func errorsCount(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "errors.count")),
		Unit:       aws.String("Count"),
		Value:      aws.Float64(float64(stats[key].Errors)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func inThroughput(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "in_throughput")),
		Unit:       aws.String("Bytes"),
		Value:      aws.Float64(float64(stats[key].InThroughput)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func outThroughput(
	stats map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "out_throughput")),
		Unit:       aws.String("Bytes"),
		Value:      aws.Float64(float64(stats[key].OutThroughput)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func systemCPUPct(
	_ map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		cpuPercent = []float64{0.0}
	}
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "system/cpu.pct")),
		Unit:       aws.String("Percent"),
		Value:      aws.Float64(cpuPercent[0]),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func systemCPUCores(
	_ map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "system/cpu_cores")),
		Unit:       aws.String("Count"),
		Value:      aws.Float64(float64(runtime.NumCPU())),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func systemMemoryAvailable(
	_ map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	// ignore error, just report 0
	mv, _ := memvalues.GetMemValues()
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "system/memory/available")),
		Unit:       aws.String("Bytes"),
		Value:      aws.Float64(float64(mv.SystemMemoryAvailable)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func systemMemoryUsed(
	_ map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	// ignore error, just report 0
	mv, _ := memvalues.GetMemValues()
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "system/memory/used")),
		Unit:       aws.String("Bytes"),
		Value:      aws.Float64(float64(mv.SystemMemoryUsed)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func systemMemoryUsedPercent(
	_ map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	// ignore error, just report 0
	mv, _ := memvalues.GetMemValues()
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "system/memory/used_percent")),
		Unit:       aws.String("Percent"),
		Value:      aws.Float64(mv.SystemMemoryUsedPercent),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}

func processMemoryUsed(
	_ map[string]grpcobserver.APIEndpointStats,
	dimensions []*cloudwatch.Dimension,
	key string,
	timestamp time.Time,
) *cloudwatch.MetricDatum {
	// ignore error, just report 0
	mv, _ := memvalues.GetMemValues()
	return &cloudwatch.MetricDatum{
		MetricName: aws.String(fmt.Sprintf("%s/%s", key, "process/memory/used")),
		Unit:       aws.String("Bytes"),
		Value:      aws.Float64(float64(mv.ProcessMemoryUsed)),
		Dimensions: dimensions,
		Timestamp:  aws.Time(timestamp),
	}
}
