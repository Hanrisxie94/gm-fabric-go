// The cloudobserver package creates a cloudObs struct that can be used to publish fm-fabric-go metrics to AWS CloudWatch.
// It also contains various helper functions to assist in the configurability and ease of creation of those cloudObs options.
package cloudobserver

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/deciphernow/gm-fabric-go/metrics/grabmetrics"
	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
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

// This is a helper function for parsing a string of dimensions.
// After checking that the string is in a parsable pattern with CheckDimStringIntegrity,
// the function gleans dimension names and values from that string
// and returns them as an array of *cloudwatch.Dimension objects.
// An empty string is allowed and will result in the metrics being stored under
// "Metrics with no dimensions" (the AWS default category for metrics with no dimensions).
// Both an empty string and a faulty string will return an empty dimensions array,
// however a faulty string will also result in an error value not equal to nil.
func ParseDimsFromString(dims string) ([]*cloudwatch.Dimension, error) {
	var dimensions []*cloudwatch.Dimension

	matched := CheckDimStringIntegrity(dims)
	if matched == false && dims != "" {
		err := errors.New("regex pattern matching error")
		errors.Wrap(err, "The given dimensions string is not formatted correctly and can't be processed.  CloudWatch metrics reporting will be aborted.")
		return dimensions, err
	}

	divideByComma := strings.Split(dims, ",")
	for _, dim := range divideByComma {
		dim = strings.TrimSpace(dim)
		divideByColon := strings.Split(dim, ":")
		dimName := strings.TrimSpace(divideByColon[0])
		dimVal := strings.TrimSpace(divideByColon[1])
		cwDim := NewDim(dimName, dimVal)
		dimensions = append(dimensions, cwDim)
	}

	return dimensions, nil
}

// This is a  helper function meant to be used in ParseDimsFromString
// in order to make sure the configured dimensions string can be parsed correctly.
// It returns a bool value of True if the string follows this kind of pattern:
// "DimensionName1: dimensionValue1, DimensionName2: dimensionValue2, ..., DimensionNameX: dimensionValueX".
// The pattern ignores extraneous whitespace around the ":" and "," delimiters,
// and purposely does not accept the special characters "@", "#", and "*"
// because they have special meanings in AWS pipeline definitions.
func CheckDimStringIntegrity(dims string) bool {
	validPattern := regexp.MustCompile(`^\s*([a-zA-Z0-9_./&+-]+\s*\:\s*[a-zA-Z0-9_./&+-]+)\s*(\,\s*[a-zA-Z0-9_./&+-]+\:\s*[a-zA-Z0-9_./&+-]+)*$`)
	return validPattern.MatchString(dims)
}

// A helper function to choose a non default method of AWS session connection
// based on which of the possible session-starting variables are provided.
// If a static credentials object is provided (can be built with the CreateStaticCreds function),
// this will be the function's first choice of variables from which to create a session.
// If a profile name is provided (and static credentials are not), the function will attempt
// to create a session based on that profile name if it's found in a SharedConfig file.
// If all else fails, the function will attempt to create a session based on an aws shared config file.
// Along with an actual session, the function will return a string tag
// that indicates what type of session was created ("default"", "static", or "profile")
// and an error message if applicable.
func ChooseSessionType(awsRegion string, awsProfile string, staticCreds *credentials.Credentials) (*session.Session, string, error) {
	awsRegion = strings.TrimSpace(awsRegion)
	awsProfile = strings.TrimSpace(awsProfile)

	defaultSess := session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))

	creds, akErr := staticCreds.Get()
	if akErr == nil {
		if len(awsRegion) != 0 {
			if len(creds.AccessKeyID) != 0 && len(creds.SecretAccessKey) != 0 {
				staticSess := session.New(&aws.Config{
					Region:      aws.String(awsRegion),
					Credentials: staticCreds,
				})
				return staticSess, "static", nil
			}
		} else {
			regionErr := errors.New("no region specified")
			errors.Wrap(regionErr, "No region was specified for creating an AWS session.  Attempting to create a session based on a SharedConfigState.")
		}
	} else {
		errors.Wrap(akErr, "The static credential variables (the AWS access key id, AWS secret access key, and AWS session token) were not set correctly.")
	}

	if len(os.Getenv("AWS_CONFIG_FILE")) == 0 {
		trueErr := errors.New("No aws config file set.")
		errors.Wrap(trueErr, "Cannot start a session with a hardcoded profile or the currently set shared config default profile.  This can't be done because the AWS_CONFIG_FILE variable is not set on this computer, so the default session will not work.")
		return defaultSess, "", trueErr
	}

	if len(awsProfile) != 0 {
		profileSess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Profile:           awsProfile,
		}))
		return profileSess, "profile", nil
	}

	err := errors.New("Unable to create a custom session based on input recieved.")
	errors.Wrap(err, "The combination of credential variables recieved was not sufficient to create an AWS session.  Attempting to start a session based on a SharedConfigState.")
	return defaultSess, "default", nil

}

// A helper function that will return a new AWS *credentials.Credentials object based on three strings:
// an aws access key id, an aws secret access key, and an aws session token.
func CreateStaticCreds(awsAccessKeyId string, awsSecretAccessKey string, awsSessionToken string) *credentials.Credentials {
	awsAccessKeyId = strings.TrimSpace(awsAccessKeyId)
	awsSecretAccessKey = strings.TrimSpace(awsSecretAccessKey)
	awsSessionToken = strings.TrimSpace(awsSessionToken)

	staticCreds := credentials.NewStaticCredentials(
		awsAccessKeyId,
		awsSecretAccessKey,
		awsSessionToken,
	)

	return staticCreds
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
		if name != "" {
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
				fmt.Println(name, "is not currently handled as a gm-fabric-go cloudwatch metric.  Skipping to the next one.")
			}
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
