// grabmetrics is a package of functions that will make it easier to grab the value of a specific metric
package grabmetrics

import (
	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/memvalues"
	"github.com/pkg/errors"
)

// LatencyP50Val is the p50 latency grabbed from a GRPCObserver
func LatencyP50Val(grpcObs *grpcobserver.GRPCObserver) float64 {
	stats, err := grpcObs.GetLatencyStats()
	if err != nil {
		errors.Wrap(err, "error getting all/latency_ms.p50")
	}
	return float64(stats["all"].P50)
}

// LatencyP95Val is the p95 latency grabbed from a GRPCObserver
func LatencyP95Val(grpcObs *grpcobserver.GRPCObserver) float64 {
	stats, err := grpcObs.GetLatencyStats()
	if err != nil {
		errors.Wrap(err, "error getting all/latency_ms.p95")
	}
	return float64(stats["all"].P95)
}

// MemUsedPercentVal is the percentage of the system's used memory
func MemUsedPercentVal() float64 {
	memValues, err := memvalues.GetMemValues()
	if err != nil {
		errors.Wrap(err, "error getting system/memory/used_percent")
	}
	return memValues.SystemMemoryUsedPercent
}

// InThroughputVal is the in throughput value grabbed from the GRPCObserver
func InThroughputVal(grpcObs *grpcobserver.GRPCObserver) float64 {
	stats, err := grpcObs.GetLatencyStats()
	if err != nil {
		errors.Wrap(err, "error getting all/in_throughput")
	}
	return float64(stats["all"].InThroughput)
}

// ErrorsCountVal is the number of errors; grabbed from the GRPCObserver
func ErrorsCountVal(grpcObs *grpcobserver.GRPCObserver) float64 {
	stats, err := grpcObs.GetLatencyStats()
	if err != nil {
		errors.Wrap(err, "error getting all/errors.count")
	}
	return float64(stats["all"].Errors)
}

// TotalRequestsVal is the number of total requests; grabbed from the GRPCObserver
func TotalRequestsVal(grpcObs *grpcobserver.GRPCObserver) float64 {
	return float64(grpcObs.GetCumulativeCount())
}
