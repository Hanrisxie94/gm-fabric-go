#!/bin/bash

set -euxo pipefail

cp test_zookeeper.yml $HOME/fabric_test_dir/src/testdir/test_service/.
pushd $HOME/fabric_test_dir/src/testdir/test_service

(
    GOPATH="${HOME}/fabric_test_dir"
    ./build_test_service_docker_image.sh
)

cat << SETTINGS > docker/settings.toml
# grpc-server
    grpc_use_tls = false
    grpc_server_host = ""
    grpc_server_port = 10000

# metrics-server
   	metrics_use_tls = false
    metrics_server_host =  ""
    metrics_server_port = 10001
    metrics_cache_size =  1024
    metrics_uri_path = "/metrics"

# gateway-proxy
	gateway_use_tls = false
    use_gateway_proxy = "true"
    gateway_proxy_host = ""
    gateway_proxy_port = 8080

# tls
    ca_cert_path = ""
    server_cert_path = ""
    server_key_path = ""
	server_cert_name = ""

# oauth
    use_oauth = false
    oauth_provider = "http://127.0.0.1:5556/dex"
    oauth_client_id = "example-app"

# zookeeper
    use_zk = true
    zk_connection_string = "zk:2181"
    zk_announce_path="/services/fabric-service/1.0"
    zk_announce_host = "127.0.0.1"

# statsd
    report_statsd = false
    statsd_host = "127.0.0.1"
    statsd_port = 8125
    statsd_mem_interval = ""

# misc
    verbose_logging = true
SETTINGS

docker-compose -f test_zookeeper.yml down
docker-clean
docker-compose -f test_zookeeper.yml  up --build 2>&1 | tee ~/zookeeper-test.log

popd
