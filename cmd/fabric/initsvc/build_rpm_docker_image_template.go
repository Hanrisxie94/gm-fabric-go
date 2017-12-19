// Copyright 2017 Decipher Technology Studios LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package initsvc

var buildRPMDockerImageTemplate = `
FROM centos:6

# Go & basic build tools
RUN yum update -y && \
    yum groupinstall -y 'Development Tools' && \
    yum install -y cyrus-sasl-devel openssl-devel libffi readline-devel && \
    curl -LO https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz && \
    tar -C /usr/local -xvzf go1.8.3.linux-amd64.tar.gz && \
    rm go1.8.3.linux-amd64.tar.gz

# Install Ruby (for fpm)
RUN git clone https://github.com/rbenv/ruby-build.git && \
    cd ruby-build && \
    ./install.sh && \
    cd .. && \
    rm -rf ruby-build && \
    ruby-build 2.3.1 /usr/local

# Install fpm
RUN gem install fpm
`
