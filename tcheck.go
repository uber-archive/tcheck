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

const (
	_serviceName = "tcheck"

	_exitUnknown          = 1
	_exitUsage            = 2
	_exitUnknownUnhealthy = 3
	_exitExplitiUnhealthy = 4
)

var _osExit = os.Exit

type exitError struct {
	code int
	msg  string
}

func (e exitError) Error() string {
	return e.msg
}

var (
	serviceName = flag.String("serviceName", "", "service name to check health of")
	peer        = flag.String("peer", "", "peer to hit directly")
)

func main() {
	flag.Parse()

	if err := healthCheck(*peer, *serviceName); err != nil {
		fmt.Println(err)
		_osExit(getExitCode(err))
	}

	fmt.Println("OK")
}

func getExitCode(err error) int {
	if ee, ok := err.(exitError); ok {
		return ee.code
	}
	return _exitUnknown
}

func healthCheck(peer, serviceName string) error {
	if peer == "" {
		return exitError{_exitUsage, "Must specify a peer to health check"}
	}
	if serviceName == "" {
		return exitError{_exitUsage, "Must specify a service name for the destination"}
	}

	ch, err := tchannel.NewChannel(_serviceName, nil)
	if err != nil {
		return err
	}

	ch.Peers().Add(peer)
	thriftClient := thrift.NewClient(ch, serviceName, nil)
	client := meta.NewTChanMetaClient(thriftClient)

	ctx, cancel := thrift.NewContext(time.Second)
	defer cancel()

	val, err := client.Health(ctx)
	if err != nil {
		return exitError{_exitUnknownUnhealthy, fmt.Sprintf("NOT OK %v\nError: %v\n", serviceName, err)}
	}
	if val.Ok != true {
		return exitError{_exitExplitiUnhealthy, fmt.Sprintf("NOT OK %v\n", *val.Message)}
	}

	return nil
}
