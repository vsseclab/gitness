// Copyright 2023 Harness, Inc.
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

package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"
)

const (
	// DefaultReadHeaderTimeout defines the default timeout for reading headers.
	DefaultReadHeaderTimeout = 2 * time.Second
)

// Config defines the config of an http server.
// TODO: expose via options?
type Config struct {
	Acme              bool
	Port              int
	Cert              string
	Key               string
	AcmeHost          string
	ReadHeaderTimeout time.Duration
}

// Server is a wrapper around http.Server that exposes different async ListenAndServe methods
// that return corresponding ShutdownFunctions.
type Server struct {
	config  Config
	handler http.Handler
}

// ShutdownFunction defines a function that is called to shutdown the server.
type ShutdownFunction func(context.Context) error

func NewServer(config Config, handler http.Handler) *Server {
	if config.ReadHeaderTimeout == 0 {
		config.ReadHeaderTimeout = DefaultReadHeaderTimeout
	}

	return &Server{
		config:  config,
		handler: handler,
	}
}

// ListenAndServe initializes a server to respond to HTTP network requests.
func (s *Server) ListenAndServe() (*errgroup.Group, ShutdownFunction) {
	if s.config.Acme {
		return s.listenAndServeAcme()
	} else if s.config.Key != "" {
		return s.listenAndServeTLS()
	}
	return s.listenAndServe()
}

func (s *Server) listenAndServe() (*errgroup.Group, ShutdownFunction) {
	var g errgroup.Group
	s1 := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.config.Port),
		ReadHeaderTimeout: s.config.ReadHeaderTimeout,
		Handler:           s.handler,
	}
	g.Go(func() error {
		return s1.ListenAndServe()
	})

	return &g, s1.Shutdown
}

func (s *Server) listenAndServeTLS() (*errgroup.Group, ShutdownFunction) {
	var g errgroup.Group
	s1 := &http.Server{
		Addr:              ":http",
		ReadHeaderTimeout: s.config.ReadHeaderTimeout,
		Handler:           http.HandlerFunc(redirect),
	}
	s2 := &http.Server{
		Addr:              ":https",
		ReadHeaderTimeout: s.config.ReadHeaderTimeout,
		Handler:           s.handler,
	}
	g.Go(func() error {
		return s1.ListenAndServe()
	})
	g.Go(func() error {
		return s2.ListenAndServeTLS(
			s.config.Cert,
			s.config.Key,
		)
	})

	return &g, func(ctx context.Context) error {
		var sg errgroup.Group
		sg.Go(func() error {
			return s1.Shutdown(ctx)
		})
		sg.Go(func() error {
			return s2.Shutdown(ctx)
		})
		return sg.Wait()
	}
}

func (s Server) listenAndServeAcme() (*errgroup.Group, ShutdownFunction) {
	var g errgroup.Group
	m := &autocert.Manager{
		Cache:      autocert.DirCache(".cache"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.config.AcmeHost),
	}
	s1 := &http.Server{
		Addr:              ":http",
		ReadHeaderTimeout: s.config.ReadHeaderTimeout,
		Handler:           m.HTTPHandler(nil),
	}
	s2 := &http.Server{
		Addr:              ":https",
		Handler:           s.handler,
		ReadHeaderTimeout: s.config.ReadHeaderTimeout,
		TLSConfig: &tls.Config{
			MinVersion:     tls.VersionTLS12,
			GetCertificate: m.GetCertificate,
			NextProtos:     []string{"h2", "http/1.1"},
		},
	}
	g.Go(func() error {
		return s1.ListenAndServe()
	})
	g.Go(func() error {
		return s2.ListenAndServeTLS("", "")
	})

	return &g, func(ctx context.Context) error {
		var sg errgroup.Group
		sg.Go(func() error {
			return s1.Shutdown(ctx)
		})
		sg.Go(func() error {
			return s2.Shutdown(ctx)
		})
		return sg.Wait()
	}
}

func redirect(w http.ResponseWriter, req *http.Request) {
	// TODO: in case of reverse-proxy the host might be not the external host.
	target := "https://" + req.Host + "/" + strings.TrimPrefix(req.URL.Path, "/")
	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}
