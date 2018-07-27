# Changelog

## 0.1.9 (July 27, 2018)

### Added

- Full ACL filter integrated into the impersonation package

### Changed

- GRPC metrics fanout handler (capture both dashboard and prometheus GRPC metrics)

## 0.1.8 (July 16th, 2018)

### Fixed

- No status messages for an empty string CloudWatch metric input

### Added

- Impersonation support for middle-man proxy
- Discovery package for Envoy management server communication

## 0.1.7 (July 11th, 2018)

### Added

- Prometheus metrics extensions

## 0.1.6 (July 2nd, 2018)

### Added
- Add capability to emit metrics to CloudWatch

## 0.1.5 (June 11th, 2018)

### Added
- Show off Go Report Card
- make ProtoFileName available to templates
- add protoc-includes config option
- whitelist http middleware

### Fixed
- Various fixes to Prometheus metrics
- Add a version identifier to the metrics dashboard JSON stream

## 0.1.4 (March 14th, 2018)

### Added
- Enable fabric generator to get template URL from config file (in addition to commandline) (#148)

### Fixed
- (metrics) improve keys in statsd display (#127)
- (metrics) eliminate slashes in functions names (#150)
- (fabric generator) Fix errors uncovered processing the swagger petstore api
- (fabric generator) Force a gRPC array to fit an JSON anonymous array
- (metrics) count events by HTTP status (#172)
- (metrics) use a function to calculate the key for HTTP metrics
- validate fabric generator version against templates
