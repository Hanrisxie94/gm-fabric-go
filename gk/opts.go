// Copyright 2018 Decipher Technology Studios LLC
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

package gk

import "crypto/tls"

// Opt is the function which will modify the options struct
type Opt func(*Options)

// Options struct holds config for service announcement
type Options struct {
	TLS *tls.Config
}

// WithTLS will add a tls object to the dialer when connecting with zookeeper
func WithTLS(cfg *tls.Config) Opt {
	return func(o *Options) {
		o.TLS = cfg
	}
}
