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

	"github.com/uber/tcheck/gen-go/meta"

	"github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/thrift"
)

func main() {
	ch, err := tchannel.NewChannel("tcheck", nil)
	if err != nil {
		fatal(1, "Failed to create tchannel: %v", err)
	}

	var (
		serviceName = flag.String("serviceName", "", "service name to check health of")
		peer        = flag.String("peer", "", "peer to hit directly")
	)
	flag.Parse()

	if *peer == "" {
		fatal(1, "Must specify a peer to health check")
	}

	ch.Peers().Add(*peer)
	thriftClient := thrift.NewClient(ch, *serviceName, nil)
	client := meta.NewTChanMetaClient(thriftClient)

	ctx, cancel := thrift.NewContext(time.Second)
	defer cancel()

	val, err := client.Health(ctx)
	if err != nil {
		fatal(2, "NOT OK %v\nError: %v\n", *serviceName, err)
	}
	if val.Ok != true {
		fatal(3, "NOT OK %v\n", *val.Message)
	}

	fmt.Printf("OK\n")
}

func fatal(code int, msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
	os.Exit(code)
}
