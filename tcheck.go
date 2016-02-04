// Copyright (c) 2016 Uber Technologies, Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/hyperbahn"
	"github.com/uber/tchannel-go/thrift"

	"github.com/uber/tcheck/gen-go/meta"
)

func main() {
	ch, err := tchannel.NewChannel("tcheck", nil)
	if err != nil {
		fmt.Println("failed to create tchannel:", err)
		os.Exit(1)
	}

	hostsFile := flag.String("hostsFile", "/etc/uber/hyperbahn/hosts.json", "hyperbahn hosts file")
	serviceName := flag.String("serviceName", "hyperbahn", "service name to check health of")
	peer := flag.String("peer", "", "peer to hit directly")

	flag.Parse()

	var config hyperbahn.Configuration
	if *peer != "" {
		config = hyperbahn.Configuration{InitialNodes: []string{*peer}}
	} else {
		config = hyperbahn.Configuration{InitialNodesFile: *hostsFile}
	}
	hyperbahn.NewClient(ch, config, nil)

	thriftClient := thrift.NewClient(ch, *serviceName, nil)
	client := meta.NewTChanMetaClient(thriftClient)

	ctx, cancel := thrift.NewContext(time.Second)
	defer cancel()

	val, err := client.Health(ctx)

	if err != nil {
		fmt.Printf("NOT OK %v\nError: %v\n", *serviceName, err)
		os.Exit(2)
	} else if val.Ok != true {
		fmt.Printf("NOT OK %v\n", *val.Message)
		os.Exit(3)
	} else {
		fmt.Printf("OK\n")
	}
}
