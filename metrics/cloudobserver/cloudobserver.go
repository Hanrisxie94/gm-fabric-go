// The cloudobserver package creates a cloudObs struct that can be used to publish fm-fabric-go metrics to AWS CloudWatch
package cloudobserver

import (
	"fmt"
	"sync"
	"time"

	"github.com/deciphernow/gm-fabric-go/metrics/grabmetrics"
	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// function options struct
type CloudOptions struct {
	sync.Mutex
	CWSess     *cloudwatch.CloudWatch
	GRPC       *grpcobserver.GRPCObserver
	Metrics    []string
	Dimensions []*cloudwatch.Dimension
	Namespace  string
}

// a single option of the CloudOptions struct
type CloudOption func(*CloudOptions)

// the CWSess option of CloudOptions
func CWSess(cwsess *cloudwatch.CloudWatch) CloudOption {
	return func(args *CloudOptions) {
		args.CWSess = cwsess
	}
}

// the GRPC option of CloudOptions
func GRPC(grpc *grpcobserver.GRPCObserver) CloudOption {
	return func(args *CloudOptions) {
		args.GRPC = grpc
	}
}

// the Metris option of CloudOptions
func Metrics(metrics []string) CloudOption {
	return func(args *CloudOptions) {
		args.Metrics = metrics
	}
}

// the Dimensions option of CloudOptions
func Dimensions(dimensions []*cloudwatch.Dimension) CloudOption {
	return func(args *CloudOptions) {
		args.Dimensions = dimensions
	}
}

// the Namespace option of CloudOptions
func Namespace(namespace string) CloudOption {
	return func(args *CloudOptions) {
		args.Namespace = namespace
	}
}

// The cloudObs struct can be used to define the AWS namespace and dimensions under which the defined metrics will reside
type cloudObs struct {
	sync.Mutex
	cwsess     *cloudwatch.CloudWatch
	grpc       *grpcobserver.GRPCObserver
	metrics    []string
	dimensions []*cloudwatch.Dimension
	namespace  string
}

// A helper function for defining a new AWS CloudWatch dimension
func NewDim(name string, value string) *cloudwatch.Dimension {
	dim := &cloudwatch.Dimension{
		Name:  aws.String(name),
		Value: aws.String(value),
	}
	return dim
}

// A helper function for appending all given dimensions into a list
func CloudDimensions(dims ...*cloudwatch.Dimension) []*cloudwatch.Dimension {
	var dimensions []*cloudwatch.Dimension
	for _, dim := range dims {
		dimensions = append(dimensions, dim)
	}
	return dimensions
}

// New returns an observer that feeds the go-metrics sink
func New(reportInterval time.Duration, setters ...CloudOption) subject.Observer {
	// default options
	args := &CloudOptions{
		CWSess:     cloudwatch.New(session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))),
		GRPC:       grpcobserver.New(2048),
		Metrics:    []string{""},
		Dimensions: CloudDimensions(NewDim("ServiceName", "default")),
		Namespace:  "Default",
	}

	for _, setter := range setters {
		setter(args)
	}

	obs := cloudObs{
		cwsess:     args.CWSess,
		grpc:       args.GRPC,
		metrics:    args.Metrics,
		dimensions: args.Dimensions,
		namespace:  args.Namespace,
	}

	go obs.AwsMetrics(reportInterval)
	return &obs
}

// Observe implements the Observer pattern
func (co *cloudObs) Observe(event subject.MetricsEvent) {
	co.Lock()
	defer co.Unlock()
}

// AWSMetrics will add defined metrics to AWS CloudWatch at the given interval
func (co *cloudObs) AwsMetrics(reportInterval time.Duration) {
	tickChan := time.Tick(reportInterval)
	for {
		<-tickChan
		co.AddAWSMetrics()
	}
}

// AddAWSMetrics handles gm-fabric-go metrics based on their names as they show up in the dashboard
func (co *cloudObs) AddAWSMetrics() {
	co.Lock()
	defer co.Unlock()
	for _, name := range co.metrics {
		switch {
		case name == "system/memory/used_percent":
			co.putMetric(name, grabmetrics.MemUsedPercentVal(), "Percent")

		case name == "all/errors.count":
			co.putMetric(name, grabmetrics.ErrorsCountVal(co.grpc), "Count")

		case name == "all/in_throughput":
			co.putMetric(name, grabmetrics.InThroughputVal(co.grpc), "Count")

		case name == "all/latency_ms.p50":
			co.putMetric(name, grabmetrics.LatencyP50Val(co.grpc), "Count")

		case name == "all/latency_ms.p95":
			co.putMetric(name, grabmetrics.LatencyP95Val(co.grpc), "Count")

		case name == "Total/requests":
			co.putMetric(name, grabmetrics.TotalRequestsVal(co.grpc), "Count")

		default:
			fmt.Println(name, "is not handled as a gm-fabric-go cloudwatch metric....yet.  Skipping to the next one.")
		}
	}
}

// putMetric is an abstraction of AWS's PutMetricData and does the work of sending a single metric value to AWS CloudWatch
func (co *cloudObs) putMetric(MetricName string, Value float64, Unit string) {
	_, err := co.cwsess.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String(co.namespace),
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				MetricName: aws.String(MetricName),
				Unit:       aws.String(Unit),
				Value:      aws.Float64(Value),
				Dimensions: co.dimensions,
			},
		},
	})
	if err != nil {
		errors.Wrap(err, err.Error())
		return
	}
}
