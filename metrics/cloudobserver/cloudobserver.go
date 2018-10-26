// The cloudobserver package creates a cloudObs struct that can be used to publish fm-fabric-go metrics to AWS CloudWatch.
// It also contains various helper functions to assist in the configurability and ease of creation of those cloudObs options.
package cloudobserver

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// The CWReporter struct can be used to define the AWS namespace and dimensions under which the defined metrics will reside
type CWReporter struct {
	CWClient     *cloudwatch.CloudWatch
	Getter       grpcobserver.LatencyStatsGetter
	Dimensions   []*cloudwatch.Dimension
	Namespace    string
	Logger       zerolog.Logger
	routesRegexp *regexp.Regexp
	datumFuncs   []datumFuncType
	Debug        bool
}

type sessAndType struct {
	sess     *session.Session
	sessType string
}

// A helper function for defining a new AWS CloudWatch dimension
func NewDim(name string, value string) *cloudwatch.Dimension {
	dim := &cloudwatch.Dimension{
		Name:  aws.String(name),
		Value: aws.String(value),
	}
	return dim
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
		return nil, errors.Errorf("The given dimensions string is not formatted correctly and can't be processed.  CloudWatch metrics reporting will be aborted.")
	}

	divideByComma := strings.Split(dims, ",")
	for _, dim := range divideByComma {
		dim = strings.TrimSpace(dim)
		divideByColon := strings.Split(dim, ":")
		if len(divideByColon) < 2 {
			return nil, errors.Errorf("Invalid format for divideByColon: '%s'", dim)
		}
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

// A helper function to retry running handlers.
func handlerRetry(r *request.Request) error {
	r.Handlers.Retry.Run(r)
	if r.Error != nil {
		return errors.Wrap(r.Error, "r.Handlers.Retry.Run")
	}

	r.Handlers.AfterRetry.Run(r)
	if r.Error != nil {
		return errors.Wrap(r.Error, "r.Handlers.AfterRetry.Run")
	}

	return nil
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
func ChooseSessionType(
	awsRegion string,
	awsProfile string,
	staticCreds *credentials.Credentials,
	validRegions []string,
) (*session.Session, string, error) {
	awsRegion, awsProfile, sess, sessType := presetValues(awsRegion, awsProfile)
	st := sessAndType{sess: sess, sessType: sessType}

	st, err, cont := useStaticSess(staticCreds, st, awsRegion, validRegions)
	if !cont {
		if err != nil {
			return nil, "", errors.Wrap(err, "useStaticSess")
		}
		return st.sess, st.sessType, nil
	}

	st, err, cont = tryConfigSess(st, awsProfile, awsRegion, validRegions)
	if !cont {
		if err != nil {
			return nil, "", errors.Wrap(err, "tryConfigSess")
		}
		return st.sess, st.sessType, nil
	}

	st, err, cont = noConfigExplicitlySet(st, awsProfile)
	if !cont {
		if err != nil {
			return nil, "", errors.Wrap(err, "noConfigExplicitlySet")
		}
		return st.sess, st.sessType, nil
	}

	return nil, "", errors.Errorf("Could not start AWS session based on combination of variables given.")
}

func noConfigExplicitlySet(st sessAndType, awsProfile string) (sessAndType, error, bool) {
	if len(awsProfile) == 0 {
		return st, nil, false
	}

	return st, nil, true
}

func tryConfigSess(st sessAndType, awsProfile string, awsRegion string, validRegions []string) (sessAndType, error, bool) {
	if len(os.Getenv("AWS_CONFIG_FILE")) == 0 {
		return st, nil, true
	}

	if len(awsProfile) != 0 {
		sessType := "profile"
		if len(awsRegion) != 0 {
			sess := session.Must(session.NewSessionWithOptions(session.Options{
				Config:  aws.Config{Region: aws.String(awsRegion)},
				Profile: awsProfile,
			}))
			sessType, err := ValidRegion(awsRegion, validRegions, sessType)
			if err != nil {
				return sessAndType{}, errors.Wrap(err, "ValidRegion"), false
			}
			st := sessAndType{sess: sess, sessType: sessType}
			return st, nil, false
		}
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Profile:           awsProfile,
		}))
		st := sessAndType{sess: sess, sessType: sessType}
		return st, nil, false

	}
	return st, nil, false
}

func useStaticSess(
	staticCreds *credentials.Credentials,
	st sessAndType,
	awsRegion string,
	validRegions []string,
) (sessAndType, error, bool) {
	creds, err := staticCreds.Get()
	if err != nil {
		return sessAndType{}, errors.Wrap(err, "The static credential variables (the AWS access key id, AWS secret access key, and AWS session token) were not set correctly."), false
	}
	if len(awsRegion) != 0 && len(creds.AccessKeyID) != 0 && len(creds.SecretAccessKey) != 0 {
		sess := session.New(&aws.Config{
			Region:      aws.String(awsRegion),
			Credentials: staticCreds,
		})
		sessType, err := ValidRegion(awsRegion, validRegions, "static")
		if err != nil {
			return sessAndType{}, errors.Wrap(err, "ValidRegion"), false
		}
		st := sessAndType{sess: sess, sessType: sessType}
		return st, nil, false
	}
	return st, nil, true
}

func presetValues(awsRegion string, awsProfile string) (string, string, *session.Session, string) {
	awsRegion = strings.TrimSpace(awsRegion)
	awsProfile = strings.TrimSpace(awsProfile)

	sess := session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))
	sessType := "default"

	return awsRegion, awsProfile, sess, sessType
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

// A helper function for creating an AWS session that will check if the provided region is valid.
func ValidRegion(inputRegion string, validRegions []string, sessType string) (string, error) {
	if len(inputRegion) == 0 {
		return "", errors.Errorf("No region provided.")
	}

	for _, region := range validRegions {
		if strings.ToLower(region) == strings.ToLower(inputRegion) {
			return sessType, nil
		}
	}

	return "", errors.Errorf("Input region does not match any AWS regions allowed for this operation.")
}

// New returns an object that can send metrics to AWS CloudWatch.
func New(
	cwReporter CWReporter,
	routes string,
	values string,
) (*CWReporter, error) {
	var err error

	cwReporter.Logger.Debug().
		Str("namespace", cwReporter.Namespace).
		Str("routes", routes).
		Str("values", values).
		Bool("debug", cwReporter.Debug).
		Msg("cloudobserver.New")

	cwReporter.routesRegexp, err = regexp.Compile(routes)
	if err != nil {
		return nil, errors.Wrapf(err, "regexp.Compile(%s) failed", routes)
	}

	// create a slice of normalized keys to datumFuncMap
	rawDatumKeys := strings.Split(values, ",")
	var datumKeys []string
	for _, rawDatumKey := range rawDatumKeys {
		datumKey := strings.TrimSpace(rawDatumKey)
		datumKey = strings.ToLower(datumKey)
		if datumKey == "" {
			continue
		}
		datumKeys = append(datumKeys, datumKey)
	}
	if len(datumKeys) == 0 {
		return nil, errors.Errorf("No CW data values specified '%v'", values)
	}

	// fill the slice of active data values with data funcs
	cwReporter.datumFuncs = make([]datumFuncType, len(datumKeys))
	for i := 0; i < len(datumKeys); i++ {
		datumFunc, ok := datumFuncMap[datumKeys[i]]
		if !ok {
			return nil, errors.Errorf("unknown value requested '%v' from '%v'",
				datumKeys[i], values)
		}
		cwReporter.datumFuncs[i] = datumFunc
	}

	return &cwReporter, nil
}

// ReportToCloudWatch will report metrics to AWS CloudWatch at the given interval
func (co *CWReporter) ReportToCloudWatch(reportInterval time.Duration) {
	co.Logger.Debug().Msgf("ReportToCloudWatch: interval %s", reportInterval)
	tickChan := time.Tick(reportInterval)
	for {
		<-tickChan
		if err := co.AddAWSMetrics(); err != nil {
			co.Logger.Error().AnErr("AddAWSMetrics", err).Msg("")
			continue
		}
		/*
			if co.Debug {
				if err := co.ListAWSMetrics(); err != nil {
					co.Logger.Error().AnErr("ListAWSMetrics", err).Msg("")
				}
			}
		*/
	}
}

// AddAWSMetrics handles the actual push of the metric up to cloudwatch.
func (co *CWReporter) AddAWSMetrics() error {
	stats, err := co.Getter.GetLatencyStats()
	if err != nil {
		return errors.Wrap(err, "GetLatencyStats()")
	}

KEY_LOOP:
	for key := range stats {

		if !co.routesRegexp.MatchString(key) {
			if co.Debug {
				co.Logger.Debug().Str("route", key).Msg("rejected: no match")
			}
			continue KEY_LOOP
		}

		timestamp := time.Now()

		if co.Debug {
			co.Logger.Debug().
				Str("key", key).
				Time("timestamp", timestamp).
				Float64("latency_ms.avg", stats[key].Avg).
				Float64("latency_ms.count", float64(stats[key].Count)).
				Float64("latency_ms.max", float64(stats[key].Max)).
				Float64("latency_ms.min", float64(stats[key].Min)).
				Float64("latency_ms.sum", float64(stats[key].Sum)).
				Float64("latency_ms.p50", float64(stats[key].P50)).
				Float64("latency_ms.p90", float64(stats[key].P90)).
				Float64("latency_ms.p95", float64(stats[key].P95)).
				Float64("latency_ms.p99", float64(stats[key].P99)).
				Float64("latency_ms.p9990", float64(stats[key].P9990)).
				Float64("latency_ms.p9999", float64(stats[key].P9999)).
				Float64("errors.count", float64(stats[key].Errors)).
				Float64("in_throughput", float64(stats[key].InThroughput)).
				Float64("out_throughput", float64(stats[key].OutThroughput)).
				Msg("AddAWSMetrics")
		}

		metricData := make([]*cloudwatch.MetricDatum, len(co.datumFuncs))
		for i := 0; i < len(co.datumFuncs); i++ {
			datum := co.datumFuncs[i](stats, co.Dimensions, key, timestamp)
			if err := datum.Validate(); err != nil {
				return errors.Wrapf(err, "datum %d invalid %s", i+1, datum.String())
			}
			metricData[i] = datum
		}

		output, err := co.CWClient.PutMetricData(
			&cloudwatch.PutMetricDataInput{
				Namespace:  aws.String(co.Namespace),
				MetricData: metricData,
			},
		)
		if err != nil {
			return errors.Wrap(err, "PutMetricData")
		}
		if co.Debug {
			co.Logger.Debug().Str("metric-output", output.String()).Msg("PutMetricData")
		}
	}

	return nil
}

// ListAWSMetrics lists available metrics for debugging
func (co *CWReporter) ListAWSMetrics() error {
	output, err := co.CWClient.ListMetrics(
		&cloudwatch.ListMetricsInput{
			// Dimensions: co.Dimensions,
			// MetricName: aws.String(fmt.Sprintf("%s/%s", testKey, "latency_ms.avg")),
			Namespace: aws.String(co.Namespace),
		},
	)
	if err != nil {
		return errors.Wrap(err, "ListMetricData")
	}
	co.Logger.Debug().Str("metric-output", output.String()).Msg("ListMetricData")
	fmt.Println(output.String())

	return nil
}

// GetAWSMetrics retrieves metrics for debugging
func (co *CWReporter) GetAWSMetrics() error {
	endTime := time.Now().Add(-(time.Hour))
	startTime := endTime.Add(-(time.Hour))
	returnData := true
	const testKey = "test_key"
	metric := cloudwatch.Metric{
		// The dimensions for the metric.
		Dimensions: co.Dimensions,

		// The name of the metric.
		MetricName: aws.String(fmt.Sprintf("%s/%s", testKey, "latency_ms.avg")),

		// The namespace of the metric.
		Namespace: aws.String(co.Namespace),
	}
	period := int64(300)
	stat := cloudwatch.MetricStat{

		// The metric to return, including the metric name, namespace, and dimensions.
		//
		// Metric is a required field
		Metric: &metric,

		// The period to use when retrieving the metric.
		//
		// Period is a required field
		Period: &period,

		// The statistic to return. It can include any CloudWatch statistic or extended
		// statistic.
		//
		// Stat is a required field
		Stat: aws.String("Sum"),

		// The unit to use for the returned data points.
		Unit: aws.String("Milliseconds"),
	}
	query := cloudwatch.MetricDataQuery{

		// The math expression to be performed on the returned data, if this structure
		// is performing a math expression. For more information about metric math expressions,
		// see Metric Math Syntax and Functions (http://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax)
		// in the Amazon CloudWatch User Guide.
		//
		// Within one MetricDataQuery structure, you must specify either Expression
		// or MetricStat but not both.
		Expression: nil,

		// A short name used to tie this structure to the results in the response. This
		// name must be unique within a single call to GetMetricData. If you are performing
		// math expressions on this set of data, this name represents that data and
		// can serve as a variable in the mathematical expression. The valid characters
		// are letters, numbers, and underscore. The first character must be a lowercase
		// letter.
		//
		// Id is a required field
		Id: aws.String("test_query"),

		// A human-readable label for this metric or expression. This is especially
		// useful if this is an expression, so that you know what the value represents.
		// If the metric or expression is shown in a CloudWatch dashboard widget, the
		// label is shown. If Label is omitted, CloudWatch generates a default.
		Label: nil,

		// The metric to be returned, along with statistics, period, and units. Use
		// this parameter only if this structure is performing a data retrieval and
		// not performing a math expression on the returned data.
		//
		// Within one MetricDataQuery structure, you must specify either Expression
		// or MetricStat but not both.
		MetricStat: &stat,

		// Indicates whether to return the time stamps and raw data values of this metric.
		// If you are performing this call just to do math expressions and do not also
		// need the raw data returned, you can specify False. If you omit this, the
		// default of True is used.
		ReturnData: &returnData,
	}

	output, err := co.CWClient.GetMetricData(
		&cloudwatch.GetMetricDataInput{
			// The time stamp indicating the latest data to be returned.
			//
			// For better performance, specify StartTime and EndTime values that align with
			// the value of the metric's Period and sync up with the beginning and end of
			// an hour. For example, if the Period of a metric is 5 minutes, specifying
			// 12:05 or 12:30 as EndTime can get a faster response from CloudWatch then
			// setting 12:07 or 12:29 as the EndTime.
			//
			// EndTime is a required field
			EndTime: &endTime,

			// The maximum number of data points the request should return before paginating.
			// If you omit this, the default of 100,800 is used.
			MaxDatapoints: nil,

			// The metric queries to be returned. A single GetMetricData call can include
			// as many as 100 MetricDataQuery structures. Each of these structures can specify
			// either a metric to retrieve, or a math expression to perform on retrieved
			// data.
			//
			// MetricDataQueries is a required field
			MetricDataQueries: []*cloudwatch.MetricDataQuery{&query},

			// Include this value, if it was returned by the previous call, to get the next
			// set of data points.
			NextToken: nil,

			// The order in which data points should be returned. TimestampDescending returns
			// the newest data first and paginates when the MaxDatapoints limit is reached.
			// TimestampAscending returns the oldest data first and paginates when the MaxDatapoints
			// limit is reached.
			ScanBy: aws.String("TimestampDescending"),

			// The time stamp indicating the earliest data to be returned.
			//
			// For better performance, specify StartTime and EndTime values that align with
			// the value of the metric's Period and sync up with the beginning and end of
			// an hour. For example, if the Period of a metric is 5 minutes, specifying
			// 12:05 or 12:30 as StartTime can get a faster response from CloudWatch then
			// setting 12:07 or 12:29 as the StartTime.
			//
			// StartTime is a required field
			StartTime: &startTime,
		},
	)
	if err != nil {
		return errors.Wrap(err, "GetMetricData")
	}
	co.Logger.Debug().Str("metric-output", output.String()).Msg("GetMetricData")

	return nil
}
