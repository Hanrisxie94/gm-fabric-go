# CloudWatch Observer #

The purpose of the `cloudobserver` package is to create a new `cloudobserver` that can publish gm-fabric-go metrics to AWS CloudWatch.

## Example ##

* Necessary:

  * create a `GRPCObserver` as defined in the `grpcobserver` package
  * create a `cloudobserver` as defined in this (the `cloudobserver`) package
  * add the `cloudobserver` (along with any other observers being used) to a metrics channel as defined in the `subject` package

* Quick start for getting some metrics to show up in AWS CloudWatch: 
  * create an http metrics manager and use it as defined in the `httpmetrics` package
  
  
```go
// An example that runs as is with a simple go run command
package main

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/deciphernow/gm-fabric-go/metrics/cloudobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/grpcobserver"
	"github.com/deciphernow/gm-fabric-go/metrics/subject"
	"time"

	"github.com/deciphernow/gm-fabric-go/metrics/httpmetrics"
	"log"
	"net/http"
)

func main() {
	grpcObserver := grpcobserver.New(2048)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	client := cloudwatch.New(sess)

	namespace := "GM/EC2"

	metrics := []string{"system/memory/used_percent"}

	dim1 := cloudobserver.NewDim(
		"AutoScalingGroupName",
		"test-asg",
	)
	dim2 := cloudobserver.NewDim(
		"ServiceName",
		"test-service",
	)
	dimensions := cloudobserver.CloudDimensions(dim1, dim2)

	cloudObserver := cloudobserver.New(
		time.Minute,
		cloudobserver.CWSess(client),
		cloudobserver.GRPC(grpcObserver),
		cloudobserver.Metrics(metrics),
		cloudobserver.Dimensions(dimensions),
		cloudobserver.Namespace(namespace),
	)

	metricsChan := subject.New(context.Background(), grpcObserver, cloudObserver)
	
	// Everything boyond this point is not directly relevant to the cloudobserver package
	// but is one quick way to see the cloudobserver at work.
	// With the help of the httpmetrics package, start an http server.
	
	httpm := httpmetrics.New(metricsChan)
	http.HandleFunc("/", httpm.HandlerFunc(http.HandlerFunc(func(http.ResponseWriter, *http.Request){})))
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}
```

## CloudObserver Configuration Details

#### Configuring the AWS session

CloudObserver must be used with AWS, which means the user or system must be 
logged in to the account metrics should be pushed to.  Before running, make
sure the appropriate environment variables, profiles, or roles are setup.

For more details on AWS credentials, see the [AWS GO SDK](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html)

#### Choosing the metrics to publish

Currently, there are five metrics that are set up to work without further work:
* system/memory/used_percent
* all/errors.count
* all/in_throughput
* all/latency_ms.p50
* all/latency_ms.p95
* Total/requests

These names are as they show up in the fabric dashboard for the service under the "Explorer" tab.  Each of these names are going to become a "Metric Name" in the AWS CloudWatch Metrics list (sorted under whichever namespace and dimensions are defined).

_Example:_
```
metrics := []string{"system/memory/used_percent", "all/errors.count",}
```

_If more metrics are needed:_

In github.com/deciphernow/gm-fabric-go/metrics/cloudobserver/cloudobserver.go there is a function that iterates through the array and tells AWS how to handle each string and turn it into a metric.  If the string that was input is not in that list of cases, it will simply be skipped over.

Expanding this will be addressed in the future.

#### Setting the namespace

The `namespace` variable is a string and refers to the AWS CloudWatch namespace under which the metrics will reside. In the example shown, we used `namespace := "GM/EC2"`.

 > **The custom namespace can have any name as long as it doesn't start with `AWS/`.**

This string will show up in CloudWatch's "All Metrics" tab in the "Custom Namespaces" section.  The namespace will hold all the dimensions, which in turn hold the actual metrics.  Using the values from the example shown, the metric(s) in the CloudWatch metrics section will be found at:

- All > GM/EC2 > AutoScalingGroupName, ServiceName > my-asg, my-service, system/memory/used_percent; my-asg, my-service, all/errors.count

#### Understanding AWS CloudWatch dimensions

A dimension consists of a name and a value.  Dimensions can essentially package the same metric in different boxes so the same metric can be reused for different services/autoscale groups/instances/etc.

> If the metrics are coming from, are part of an alarm for, or generally have anything to do with an AWS EC2 instance, the defined dimensions must be some combination of the four AWS EC2 metric dimensions: `AutoScalingGroupName`, `ImageId`, `InstanceId`, and `InstanceType`.  With this in mind, the above example is kind of silly because `ServiceName` is not one of the four AWS EC2 dimensions so, although a metric with these dimensions could be attached to an alarm for the `my-asg` autoscale group, the alarm will remain in an insufficient state no matter how many data points get pushed to CloudWatch.

#### Setting the collection frequency

Though CloudObserver doesn't limit the frequency of metrics publishing, AWS Cloudwatch does have some limitations that need to be understood.  From the  [AWS CloudWatch documentation](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/publishingMetrics.html):

>Although you can publish data points with time stamps as granular as one-thousandth of a second, CloudWatch aggregates the data to a minimum granularity of one minute. CloudWatch records the average (sum of all items divided by number of items) of the values received for every 1-minute period, as well as the number of samples, maximum value, and minimum value for the same time period.

The above means that even if the `reportInterval` is shorter than a minute,  
a data point will only pop up every minute.  Though the example below 
collects and publishes metrics every thirty seconds, the value shown on 
CloudWatch will be the average every minute.  

_Example:_
```go
import "time"

// report metrics every 30 seconds
reportInterval := time.Duration(30) * time.Second
```

#### Using GRPCObserver to collect metrics

The variable `grpc` is a `*grpcobserver.GRPCObserver` object.  This does the
actual metric collection from the code/service, from which CloudObserver simply
extracts the metrics that it reports to CloudWatch.  In the case that
metrics need to be collected from a different source, this could be replaced by
any other observer pattern.


