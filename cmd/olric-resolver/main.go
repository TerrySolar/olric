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

package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/buraksezer/olric"
	"github.com/buraksezer/olric/internal/process"
	"github.com/buraksezer/olric/internal/resolver"
	"github.com/buraksezer/olric/internal/resolver/config"
	"github.com/sean-/seed"
)

func usage() {
	var msg = `Usage: olric-resolver [options] ...

Transaction resolver component of Olric

Options:
  -h, --help    Print this message and exit.
  -v, --version Print the version number and exit.
  -c, --config  Sets configuration file path. Default is olric-resolver-local.yaml in the
                current folder. Set OLRIC_RESOLVER_CONFIG to overwrite it.

The Go runtime version %s
Report bugs to https://github.com/buraksezer/olric/issues
`
	_, err := fmt.Fprintf(os.Stdout, msg, runtime.Version())
	if err != nil {
		panic(err)
	}
}

type arguments struct {
	config  string
	help    bool
	version bool
}

const (
	// DefaultConfigFile is the default configuration file path on a Unix-based operating system.
	DefaultConfigFile = "olric-resolver-local.yaml"

	// EnvConfigFile is the name of environment variable which can be used to override default configuration file path.
	EnvConfigFile = "OLRIC_RESOLVER_CONFIG"
)

func main() {
	args := &arguments{}

	// Parse command line parameters
	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	f.SetOutput(ioutil.Discard)
	f.BoolVar(&args.help, "h", false, "")
	f.BoolVar(&args.help, "help", false, "")

	f.BoolVar(&args.version, "version", false, "")
	f.BoolVar(&args.version, "v", false, "")

	f.StringVar(&args.config, "config", DefaultConfigFile, "")
	f.StringVar(&args.config, "c", DefaultConfigFile, "")

	if err := f.Parse(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("Failed to parse flags: %v\n", err))
		usage()
		os.Exit(1)
	}

	if args.version {
		_, _ = fmt.Fprintf(os.Stderr, "olric-resolver version %s %s %s/%s\n",
			olric.ReleaseVersion,
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
		)
		return
	} else if args.help {
		usage()
		return
	}

	// MustInit provides guaranteed secure seeding.  If `/dev/urandom` is not
	// available, MustInit will panic() with an error indicating why reading from
	// `/dev/urandom` failed.  MustInit() will upgrade the seed if for some reason a
	// call to Init() failed in the past.
	seed.MustInit()

	c, err := config.New(args.config)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] Failed to load the configuration: %v\n", err)
		os.Exit(1)
	}

	lg := log.New(os.Stderr, "", log.LstdFlags)
	rs, err := resolver.New(c, lg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[ERROR] Failed to load the configuration: %v\n", err)
		os.Exit(1)
	}

	p, err := process.New(rs)
	if err != nil {
		lg.Fatalf("[ERROR] Failed to create a new %s instance: %v", rs.Name(), err)
	}

	if err = p.Start(); err != nil {
		lg.Printf("[ERROR] Failed to start %s: %v", rs.Name(), err)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err := p.Shutdown(ctx); err != nil {
			lg.Printf("[ERROR] Failed to shutdown %s: %v", rs.Name(), err)
		}
		lg.Fatal("[ERROR] Quit unexpectedly!")
	}

	lg.Printf("[INFO] Quit!")
}