// Copyright 2018-2022 Burak Sezer
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

package config

type Sequencer struct {
	Addr string
}

type Config struct {
	DataDir   string
	Sequencer Sequencer
}

// FIXME: This will be removed after merging into master
func DefaultConfig() *Config {
	return &Config{
		DataDir: "/Users/buraksezer/data/olric-zmap",
		Sequencer: Sequencer{
			Addr: "localhost:4545",
		},
	}
}