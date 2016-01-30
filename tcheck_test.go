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
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/testutils"
	"github.com/uber/tchannel-go/thrift"

	"github.com/stretchr/testify/assert"
)

func healthNotOk(ctx thrift.Context) (ok bool, message string) {
	return false, "hello world"
}

func healthOk(ctx thrift.Context) (ok bool, message string) {
	return true, "hello world"
}

func TestOk(t *testing.T) {
	channel, _, hostPort := SetupServer(t, healthOk)
	strOut, err := Run([]string{fmt.Sprintf("--peer=%s", hostPort), "--serviceName=testing"})

	assert.NoError(t, err, "no error from tcheck")

	assert.Equal(t, "OK\n", strOut, "tcheck correct output")

	channel.Close()
}

func TestNotOk(t *testing.T) {
	channel, _, hostPort := SetupServer(t, healthNotOk)
	strOut, err := Run([]string{fmt.Sprintf("--peer=%s", hostPort), "--serviceName=testing"})

	strErr := fmt.Sprintf("%v", err)
	assert.Equal(t, "exit status 3", strErr, "correct return code")

	assert.Equal(t, "NOT OK hello world\n", strOut, "tcheck correct output")

	channel.Close()
}

func SetupServer(t *testing.T, fn thrift.HealthFunc) (*tchannel.Channel, *thrift.Server, string) {
	_, cancel := tchannel.NewContext(time.Second * 10)
	defer cancel()

	tchan := testutils.NewServer(t, testutils.NewOpts().SetServiceName("testing"))
	server := thrift.NewServer(tchan)

	server.RegisterHealthHandler(fn)

	return tchan, server, tchan.PeerInfo().HostPort
}

func Run(args []string) (string, error) {
	cmd := exec.Command("./tcheck", args...)
	out, err := cmd.Output()
	strOut := string(out)
	return strOut, err
}
