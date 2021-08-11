// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package http

import (
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"
	"context"
	"github.com/openziti/sdk-golang/ziti"
)

// Options holds http configuration properties for communicating with InfluxDB server
type Options struct {
	// HTTP client. Default is http.DefaultClient.
	httpClient *http.Client
	// doer is an http Doer - if set it overrides httpClient
	doer Doer
	// Flag whether http client was created internally
	ownClient bool
	// TLS configuration for secure connection. Default nil
	tlsConfig *tls.Config
	// HTTP request timeout in sec. Default 20
	httpRequestTimeout uint
}

type ZitiDialContext struct {
	context ziti.Context
}

func (dc *ZitiDialContext) Dial(_ context.Context, _ string, addr string) (net.Conn, error) {
	service := strings.Split(addr, ":")[0] // will always get passed host:port
	return dc.context.Dial(service)
}

// HTTPClient returns the http.Client that is configured to be used
// for HTTP requests. It will return the one that has been set using
// SetHTTPClient or it will construct a default client using the
// other configured options.
// HTTPClient panics if SetHTTPDoer was called.
func (o *Options) HTTPClient() *http.Client {
	if o.doer != nil {
		panic("HTTPClient called after SetHTTPDoer")
	}
	if o.httpClient == nil {
		zitiDialContext := ZitiDialContext{context: ziti.NewContext()}
		zitiTransport := http.DefaultTransport.(*http.Transport).Clone() // copy default transport
		zitiTransport.DialContext = zitiDialContext.Dial
		o.httpClient = &http.Client{
			Timeout: time.Second * time.Duration(o.HTTPRequestTimeout()),
			Transport: zitiTransport,
		}
		o.ownClient = true
	}
	return o.httpClient
}

// SetHTTPClient will configure the http.Client that is used
// for HTTP requests. If set to nil, an HTTPClient will be
// generated.
//
// Setting the HTTPClient will cause the other HTTP options
// to be ignored.
// In case of UsersAPI.SignIn() is used, HTTPClient.Jar will be used for storing session cookie.
func (o *Options) SetHTTPClient(c *http.Client) *Options {
	o.httpClient = c
	o.ownClient = false
	return o
}

// OwnHTTPClient returns true of HTTP client was created internally. False if it was set externally.
func (o *Options) OwnHTTPClient() bool {
	return o.ownClient
}

// Doer allows proving custom Do for HTTP operations
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// SetHTTPDoer will configure the http.Client that is used
// for HTTP requests. If set to nil, this has no effect.
//
// Setting the HTTPDoer will cause the other HTTP options
// to be ignored.
func (o *Options) SetHTTPDoer(d Doer) *Options {
	if d != nil {
		o.doer = d
		o.ownClient = false
	}
	return o
}

// HTTPDoer returns actual Doer if set, or http.Client
func (o *Options) HTTPDoer() Doer {
	if o.doer != nil {
		return o.doer
	}
	return o.HTTPClient()
}

// TLSConfig returns tls.Config
func (o *Options) TLSConfig() *tls.Config {
	return o.tlsConfig
}

// SetTLSConfig sets TLS configuration for secure connection
func (o *Options) SetTLSConfig(tlsConfig *tls.Config) *Options {
	o.tlsConfig = tlsConfig
	return o
}

// HTTPRequestTimeout returns HTTP request timeout
func (o *Options) HTTPRequestTimeout() uint {
	return o.httpRequestTimeout
}

// SetHTTPRequestTimeout sets HTTP request timeout in sec
func (o *Options) SetHTTPRequestTimeout(httpRequestTimeout uint) *Options {
	o.httpRequestTimeout = httpRequestTimeout
	return o
}

// DefaultOptions returns Options object with default values
func DefaultOptions() *Options {
	return &Options{httpRequestTimeout: 20}
}
