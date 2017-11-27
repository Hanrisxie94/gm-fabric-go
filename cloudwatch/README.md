# Cloudwatch
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/deciphernow/gm-fabric-go/cloudwatch)

Package middleware provides a simple middleware abstraction on top of net/http.

1.  [Prerequisites](#prerequisites)
2.  [Install](#install)
3.  [Usage](#usage)

## Prerequisites

1.  [Go](https://golang.org) (1.9+ recommended)

## Installation

Using [golang/dep](https://github,com/golang/dep):
```bash
dep ensure -v -add github.com/deciphernow/gm-fabric-go/cloudwatch
```

## Usage
The cloudwatch package provides a metric publisher. To publish metrics, you'll use some
pattern similar to:
```go
package main

import (
    "github.com/deciphernow/gm-fabric-go/cloudwatch"
)

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
```
