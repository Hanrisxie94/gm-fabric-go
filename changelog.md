# Changelog

## 0.1.6

* Add capability to emit metrics to CloudWatch

## 0.1.5

* Show off Go Report Card
* Various fixes to Prometheus metrics (in progress)
* Add a version identifier to the metrics dashboard JSON stream 
* make ProtoFileName available to templates 
* add protoc-includes config option 
* whitelist http middleware 

## 0.1.4

* (metrics) improve keys in statsd display (#127)
* Enable fabric generator to get template URL from config file (in addition to commandline) (#148)
* (metrics) eliminate slashes in functions names (#150)
* (fabric generator) Fix errors uncovered processing the swagger petstore api
* (fabric generator) Force a gRPC array to fit an JSON anonymous array
* (metrics) count events by HTTP status (#172)
* (metrics) use a function to calculate the key for HTTP metrics
* validate fabric generator version against templates