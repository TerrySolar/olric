// Copyright 2018-2021 Burak Sezer
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

package resp

import (
	"github.com/buraksezer/olric/internal/util"
	"github.com/tidwall/redcon"
)

func ParsePingCommand(cmd redcon.Command) (*Ping, error) {
	if len(cmd.Args) < 1 {
		return nil, errWrongNumber(cmd.Args)
	}

	p := NewPing()
	if len(cmd.Args) == 2 {
		p.SetMessage(util.BytesToString(cmd.Args[1]))
	}
	return p, nil
}
