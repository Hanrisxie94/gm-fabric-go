# Consul Service Discovery
[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/deciphernow/gm-fabric-go/consul)

Easy consul service discovery with health-check capabilities

1.  [Prerequisites](#prerequisites)
2.  [Getting Started](#getting-started)
3.  [Info](#info)
4.  [Glossary](#glossary)

## Prerequisites

1.  We assume [Consul](https://www.consul.io/) is running somewhere
2.  [Go](https://golang.org/dl) (Recommended 1.9+)

## Getting Started
1.  Create a new client that can talk to consul. Provide the local host and port of the consul agent:
```go
c, err := consul.NewConsulClient("consul:8500")
if err != nil {
    log.Fatal(err)
}
```
2.  Create service configs which will then be used to register your service with the consul agent:
    -   We recommend reading from a config file. We urge you to read through the Glossary to see what the struct tags for each config object so we can name the fields appropriately in your config file. See supported file types: [Glossary](#glossary)
    ```toml
    [service]
    name = "my_service"
    host = "0.0.0.0"
    port = 8080

    [health_check]
    protocol = "http"
    interval = "10s"
    timeout = "10s"
    endpoint = "/api/test"
    ```
    -   Unmarshal the config file into the following objects:
    ```go
    // ServiceConf would contain other configuration options the micro service might use.
    type ServiceConf {
        SC consul.ServiceConfig `toml:"service"`
        HC consul.CheckConfig   `toml:"health_check"`
    }

    var c ServiceConf
    // We are using BurntSushi/toml to read the config file here
    if _, err := toml.DecodeFile("./config.toml", &c); err != nil {
        log.Fatal(err)
    }
    ```
3. Register your service with your newly created config objects
```go
// Register will return a unique hash that is used to register the service with consul in the agent cluster
id, err := c.Register(
    consul.WithServiceConfig(c.SC),
    consul.WithCheckConfig(c.HC),
    consul.WithTags([]string{"http", "https", "gRPC"}),
)
// Check your error. Service registration isn't guaranteed
if err != nil{
    log.Fatal(err)
}
log.Println("Successfully registered service with consul agent: " id)
```
3.  If you need to deregister for any reason, you can do the following
```go
// id is the hash we got back from register
err = c.Deregister(id)
if err != nil {
    log.Fatal(err)
}
log.Println("Successfully deregistered service from consul agent ")
```
## Info
1.  ServiceConfig (required) - configuration for basic service registration
```go
// ServiceConfig is the basic service configuration object
type ServiceConfig struct {
	Name string `json:"service_name" toml:"name" yaml:"name"`
	Host string `json:"service_host" toml:"host" yaml:"host"`
	Port int    `json:"service_port" toml:"port" yaml:"port"`
}

```
2.  CheckConfig (optional) - configuration if you wish to register your service with a health-check
```go
// CheckConfig is the object you pass to Register that provides configuration option to the Consul Service Check
type CheckConfig struct {
	Protocol string `json:"protocol" toml:"protocol" yaml:"protocol"`     // http - https - tcp
	Interval string `json:"interval" toml:"interval" yaml:"interval"`     // how often you want the health check to run
	Timeout  string `json:"timeout" toml:"timeout" yaml:"timeout"`        // how long the health check should wait before it times out
	Endpoint string `json:"api_endpoint" toml:"endpoint" yaml:"endpoint"` // service endpoint the health check will ping
}
```

## Glossary
__Tags__  - unique names you can apply to a service that serve as searchable metadata in the consul agent.

__ID__ - unique hash used to register service with consul agent. This can also be used to deregister a service.

---
Currently supported file types for direct un-marshaling:
-   YAML
-   TOML
-   JSON
