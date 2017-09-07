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
	"net"
	"os"
	"strings"
	"time"

	"github.com/uber/tcheck/internal/gen-go/meta"

	"github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/thrift"
)

//go:generate thrift-gen --generateThrift --inputFile meta.thrift --outputDir ./internal/gen-go

const (
	_serviceName = "tcheck"

	_exitUnknown           = 1
	_exitUsage             = 2
	_exitUnknownUnhealthy  = 3
	_exitExplicitUnhealthy = 4
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
	peer        = flag.String("peer", "", "Peer host:port to health check")
	serviceName = flag.String("serviceName", "", "Service name to health check")
	timeout     = flag.Duration("timeout", time.Second, "Timeout for the health check")
)

func main() {
	flag.Parse()

	if err := healthCheck(*peer, *serviceName, *timeout); err != nil {
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

func healthCheck(peer, serviceName string, timeout time.Duration) error {
	if peer == "" {
		return exitError{_exitUsage, "Must specify a peer to health check"}
	}
	if serviceName == "" {
		return exitError{_exitUsage, "Must specify a service name for the destination"}
	}
	if timeout <= 0 {
		return exitError{_exitUsage, "Must specify a positive timeout"}
	}

	ch, err := tchannel.NewChannel(_serviceName, nil)
	if err != nil {
		return err
	}

	peer = remapLocalhost(peer)
	ch.Peers().Add(peer)
	thriftClient := thrift.NewClient(ch, serviceName, nil)
	client := meta.NewTChanMetaClient(thriftClient)

	ctx, cancel := thrift.NewContext(timeout)
	defer cancel()

	val, err := client.Health(ctx)
	if err != nil {
		return exitError{_exitUnknownUnhealthy, fmt.Sprintf("NOT OK %v\nError: %v\n", serviceName, err)}
	}
	if val.Ok != true {
		return exitError{_exitExplicitUnhealthy, fmt.Sprintf("NOT OK %v\n", val.GetMessage())}
	}

	return nil
}

// TChannel tools remap the string "localhost" to the best public IP on the host.
// The protocol assumes services only listen on a single host:port, so the
// client libraries tend to listen on a single interface. This localhost remapping
// makes it easier for users to health check an instance on the current host
// without having to find the public IP of the machine.
func remapLocalhost(hostPort string) string {
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return hostPort
	}

	if !strings.EqualFold(host, "localhost") {
		return hostPort
	}

	ip, err := tchannel.ListenIP()
	if err != nil {
		return hostPort
	}

	return net.JoinHostPort(ip.String(), port)
}
