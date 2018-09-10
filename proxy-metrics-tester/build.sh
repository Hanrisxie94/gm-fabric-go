#!/bin/bash

set -euxo pipefail

docker build -f Dockerfile -t drfogout/proxy-metrics-tester .