# CloudWatch Observer #

The purpose of the `cloudobserver` package is to create a new `cloudobserver` object that can publish gm-fabric-go metrics to AWS CloudWatch.

## Example ##

* Necessary:
  * define a new `cloudobserver` using this `cloudobserver` package

* Quick start steps for getting some metrics to show up in AWS CloudWatch: 
  * add the `cloudobserver` (along with any other observers being used) to a metrics channel as defined in the `subject` package
  * create an http metrics manager and use it as defined in the `httpmetrics` package
  
  
```go
// An example that runs as is with a simple go run command
package main

import (
	"context"
	"github.com/deciphernow/gm-fabric-go/metrics/cloudobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
	"time"

	"github.com/deciphernow/gm-fabric-go/metrics/httpmetrics"
	"log"
	"net/http"
)

func main() {

	metrics := []string{"system/memory/used_percent, Total/requests"}
	dimensions, err := cloudobserver.ParseDimsFromString("ServiceName: test-service")
	if err != nil {
		return
	}
	namespace := "GM/EC2"

	// This cloudObserver makes use of the following defaults:
	// * CloudWatch session based on the default profile of the local aws config files
	// * A grpcobserver.New(2048)
	cloudObserver := cloudobserver.New(
		time.Minute,
		cloudobserver.Metrics(metrics),
		cloudobserver.Dimensions(dimensions),
		cloudobserver.Namespace(namespace),
	)

	// Everything boyond this point is not directly relevant to the cloudobserver package
	// but is one quick way to see the cloudobserver at work.
	// With the help of the httpmetrics package, 
	// start an http server using a metrics channel as defined in the `subject` package.

	metricsChan := subject.New(context.Background(), cloudObserver)
	httpm := httpmetrics.New(metricsChan)
	http.HandleFunc("/", httpm.HandlerFunc(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})))
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

```

## CloudObserver Configuration Details

A cloudobserver object can be created using the `cloudobserver.New()` function.

The `New()` function needs a `time.Duration` parameter to specify how often to publish metrics to CloudWatch.

It will optionally take any combination (or none) of the following parameters:
* `cloudobserver.CWSess()`: accepts a `*cloudwatch.CloudWatch` object
* `cloudobserver.GRPC()`: accepts a `*grpcobserver.GRPCObserver` object
* `cloudobserver.Metrics()`: accepts a `[]string` object
* `cloudobserver.Dimensions()`: accepts a `[]*cloudwatch.Dimension` object
* `cloudobserver.Namespace()`: accepts a `string` object

If the optional parameters are not set, they use a default value.  It is recommended to not rely on the default values for at least the metrics, dimensions, and namespace parameters.

### Configuring the AWS session

**Parameter:**

`cloudobserver.CWSess(input)`

**Default Input:**

`cloudwatch.New(session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))),
`
  * This means an AWS CloudWatch session will attempt to connect using the default profile as defined in the local shared config files.
  
**Example:**

```go
import (
	"github.com/deciphernow/gm-fabric-go/metrics/cloudobserver"
)

session := cloudwatch.New(session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable})))

cloudObserver := cloudobserver.New(
	time.Minute,
	cloudobserver.CWSess(session),
)
```

**More Information:**

The `cloudobserver` must be used with AWS, which means the user or system must have the credentials to log in to the account the metrics should be pushed to.  Before running, the appropriate environment variables, profiles, or roles should be set up.

For more details on AWS credentials, see the [AWS GO SDK](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html)

Some helper functions exist in the package to assist in creating the session.  Most notably, `ChooseSessionType` can take in a variety of parameters that _could_ be used to start a session, and choose the most appropriate method from the input recieved to output an AWS `*session.Session` object (or throw an error explaining why the session can't be created).

### Choosing the metrics to publish

**Parameter:**

`cloudobserver.Metrics(input)`

**Default Input:**

`[]string{""}`

**Example:**

[_above_](#example)

**More Information:**

Currently, there are five metrics that are set up to be published to AWS:
* system/memory/used_percent
* all/errors.count
* all/in_throughput
* all/latency_ms.p50
* all/latency_ms.p95
* Total/requests

These names are as they show up in the fabric dashboard for the service under the "Explorer" tab.  Each of these names are going to become a "Metric Name" in the AWS CloudWatch Metrics list (sorted under whichever namespace and dimensions are defined).

_If more metrics are needed:_

In github.com/deciphernow/gm-fabric-go/metrics/cloudobserver/cloudobserver.go there is a function that iterates through the array and tells AWS how to handle each string and turn it into a metric.  If the string that was input is not in that list of cases, it will simply be skipped over.

Expanding this will be addressed in the future.

### Setting the namespace

**Parameter:**

`cloudobserver.Namespace(input)`

**Default Input:**

`"Default"`

**Example:**

[_above_](#example)

**More Information:**

The `Namespace` variable is a string and refers to the AWS CloudWatch namespace under which the metrics will reside.

 > **The custom namespace can have any name as long as it doesn't start with `AWS/`.**

This string will show up in CloudWatch's "All Metrics" tab in the "Custom Namespaces" section.  The namespace will hold all the dimensions, which in turn hold the actual metrics.  Using the values from the example shown, the metric(s) in the CloudWatch metrics section will be found at:

- All > GM/EC2 > AutoScalingGroupName, ServiceName > my-asg, my-service, system/memory/used_percent; my-asg, my-service, all/errors.count

### Understanding AWS CloudWatch dimensions

**Parameter:**

`cloudobserver.Dimensions(input)`

**Default Input:**

`CloudDimensions(NewDim("ServiceName", "default"))`

  * This is setting a single dimension with the name `ServiceName` and the associated value of `default`.
	
**More Information:**

A dimension consists of a name and a value.  Dimensions can essentially package the same metric in different boxes so the same metric can be reused for different services/autoscale groups/instances/etc.

> If the metrics are coming from, are part of an alarm for, or generally have anything to do with an AWS EC2 instance, the defined dimensions must be some combination of the four AWS EC2 metric dimensions: `AutoScalingGroupName`, `ImageId`, `InstanceId`, and `InstanceType`.  With this in mind, the above example is kind of silly because `ServiceName` is not one of the four AWS EC2 dimensions so, although a metric with these dimensions could be attached to an alarm for the `my-asg` autoscale group, the alarm will remain in an insufficient state no matter how many data points get pushed to CloudWatch.

The `cloudobserver` package offers a function, `ParseDimsFromString` that takes in an appropriately formatted string and returns  a `[]*cloudwatch.Dimension` object.

### Setting the collection frequency

**Parameter:**

A `time.Duration` object.

**Default:**

No default.  Must be explicitly set.

**Example:**

[_above_](#example)

**More Information:**

Though CloudObserver doesn't limit the frequency of metrics publishing, AWS Cloudwatch does have some limitations that need to be understood.  From the  [AWS CloudWatch documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/publishingMetrics.html):

>Although you can publish data points with time stamps as granular as one-thousandth of a second, CloudWatch aggregates the data to a minimum granularity of one minute. CloudWatch records the average (sum of all items divided by number of items) of the values received for every 1-minute period, as well as the number of samples, maximum value, and minimum value for the same time period.

The above means that even if the `reportInterval` is shorter than a minute, a data point will only pop up every minute.  If the interval is set to publish metrics every 15 seconds, the value available on CloudWatch will be the average (or min, max, etc.) of four data points every minute.

### Using GRPCObserver to collect metrics

**Parameter:**

`cloudobserver.GRPC(input)`

**Default Input:**

`grpcobserver.New(2048)`

**Example:**

``` go
import (
	"github.com/deciphernow/gm-fabric-go/metrics/cloudobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"
)

grpc := grpcobserver.New(2048)

cloudObserver := cloudobserver.New(
	time.Minute,
	cloudobserver.GRPC(grpc),
)
```

**More Information:**

This parameter needs a `*grpcobserver.GRPCObserver` object.  As of the creation of this package, the defined Grey Matter metrics configured for use by AWS were most easily collected via a GRPC observer.

Expanding this will be addressed in the future.  It could include collecting the metrics from Prometheus, or abstracting this parameter to accept any type of observer (or multiple types of observers).
