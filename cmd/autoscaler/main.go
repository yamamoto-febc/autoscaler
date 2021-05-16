// Copyright 2021 The sacloud Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// AutoScaler Core
//
// Usage:
//   autoscaler [flags]
//
// Flags:
//   -address: (optional) URL of gRPC endpoint of AutoScaler Core. default:`unix:autoscaler.sock`
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/sacloud/autoscaler/core"
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/request"
	"github.com/sacloud/autoscaler/version"
	"google.golang.org/grpc"
)

func main() {
	var address string
	flag.StringVar(&address, "address", defaults.CoreSocketAddr, "URL of gRPC endpoint of AutoScaler Core")

	var showHelp, showVersion bool
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")

	flag.Parse()

	// TODO add flag validation

	switch {
	case showHelp:
		showUsage()
		return
	case showVersion:
		fmt.Println(version.FullVersion())
		return
	default:
		// TODO 簡易的な実装、後ほど整理&切り出し
		filename := strings.Replace(defaults.CoreSocketAddr, "unix:", "", -1)
		lis, err := net.Listen("unix", filename)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			lis.Close()
			if _, err := os.Stat(filename); err == nil {
				if err := os.RemoveAll(filename); err != nil {
					log.Fatal(err)
				}
			}
		}()

		log.Printf("autoscaler started with: %s\n", lis.Addr().String())

		server := grpc.NewServer()
		srv := core.NewScalingService()
		request.RegisterScalingServiceServer(server, srv)

		if err := server.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}
}

func showUsage() {
	fmt.Println("usage: autoscaler [flags]")
	flag.Usage()
}
