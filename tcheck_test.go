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
	"errors"
	"os"
	"runtime"
	"testing"

	"github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/testutils"
	"github.com/uber/tchannel-go/thrift"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func healthNotOk(ctx thrift.Context) (ok bool, message string) {
	return false, "hello world"
}

func healthOk(ctx thrift.Context) (ok bool, message string) {
	return true, "hello world"
}

func setupServer(t *testing.T, fn thrift.HealthFunc) *tchannel.Channel {
	opts := testutils.NewOpts().
		SetServiceName("svc").
		DisableLogVerification()
	tchan := testutils.NewServer(t, opts)

	if fn != nil {
		server := thrift.NewServer(tchan)
		server.RegisterHealthHandler(fn)
	}
	return tchan
}

func TestHealthCheckBadArgs(t *testing.T) {
	noHandler := setupServer(t, nil)
	defer noHandler.Close()

	unhealthyHandler := setupServer(t, func(_ thrift.Context) (ok bool, msg string) {
		return false, "test-error"
	})
	defer unhealthyHandler.Close()

	healthyHandler := setupServer(t, func(_ thrift.Context) (ok bool, msg string) {
		return true, ""
	})

	tests := []struct {
		msg      string
		peer     string
		svc      string
		fn       thrift.HealthFunc
		wantExit int
		wantErr  string
	}{
		{
			msg:      "missing service",
			peer:     "127.0.0.1",
			svc:      "",
			wantExit: _exitUsage,
		},
		{
			msg:      "missing peer",
			peer:     "",
			svc:      "svc",
			wantExit: _exitUsage,
		},
		{

			msg:      "healthy server",
			peer:     healthyHandler.PeerInfo().HostPort,
			svc:      "svc",
			wantExit: 0,
		},
		{
			msg:      "no health handler",
			peer:     noHandler.PeerInfo().HostPort,
			svc:      "svc",
			wantExit: _exitUnknownUnhealthy,
			wantErr:  "ErrCodeBadRequest",
		},
		{
			msg:      "unhealthy health handler",
			peer:     unhealthyHandler.PeerInfo().HostPort,
			svc:      "svc",
			wantExit: _exitExplitiUnhealthy,
			wantErr:  "test-error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			err := healthCheck(tt.peer, tt.svc)
			if tt.wantExit > 0 {
				require.Error(t, err)
				assert.Equal(t, tt.wantExit, getExitCode(err), "Unexpected error code")
				assert.Contains(t, err.Error(), tt.wantErr, "Missing expected error")
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestIntegrationSuccess(t *testing.T) {
	server := setupServer(t, nil)
	defer server.Close()

	// Set up a default health handler.
	thrift.NewServer(server)

	os.Args = []string{"tcheck", "--peer", server.PeerInfo().HostPort, "--serviceName", server.ServiceName()}
	main()
}

func TestIntegrationError(t *testing.T) {
	defer func() { _osExit = os.Exit }()

	var exitCode int
	_osExit = func(code int) {
		exitCode = code
		runtime.Goexit()
	}

	server := setupServer(t, nil)
	defer server.Close()

	// Start a separate goroutine for the main function since we stub out _osExit
	// to kill the current goroutine.
	done := make(chan struct{})
	go func() {
		defer close(done)

		os.Args = []string{"tcheck", "--peer", server.PeerInfo().HostPort, "--serviceName", server.ServiceName()}
		main()
	}()

	// Wait for the main function to end.
	<-done
	assert.Equal(t, _exitUnknownUnhealthy, exitCode, "Expected non-zero exit")
}

func TestGetExitCode(t *testing.T) {
	assert.Equal(t, 5, getExitCode(exitError{5, ""}))
	assert.Equal(t, 1, getExitCode(errors.New("unknown")))
}
