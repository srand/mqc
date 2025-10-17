package transport

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/srand/mqc"
)

type TransportOptions struct {
	Addrs []string

	// Timeout for the dial operation
	ConnectTimeout time.Duration // in seconds

	// Timeout for individual RPC calls
	CallTimeout time.Duration // in seconds

	// Underlying protocol to use (e.g. "tcp", "unix").
	// Varies depending on the transport implementation.
	Protocol string

	// TLS configuration for secure connections
	TlsConfig *tls.Config

	// OnConnect is a callback function that is called when a connection is established.
	OnConnect func(mqc.Transport)
}

type TransportOption func(*TransportOptions) error

func WithAddress(addr string) TransportOption {
	return func(opts *TransportOptions) error {
		opts.Addrs = append(opts.Addrs, addr)
		return nil
	}
}

func WithConnectTimeout(d time.Duration) TransportOption {
	return func(opts *TransportOptions) error {
		opts.ConnectTimeout = d
		return nil
	}
}

func WithCallTimeout(d time.Duration) TransportOption {
	return func(opts *TransportOptions) error {
		opts.CallTimeout = d
		return nil
	}
}

func WithProtocol(protocol string) TransportOption {
	return func(opts *TransportOptions) error {
		opts.Protocol = protocol
		return nil
	}
}

func WithTLSConfig(tlsConfig *tls.Config) TransportOption {
	return func(opts *TransportOptions) error {
		opts.TlsConfig = tlsConfig
		return nil
	}
}

func WithCertificateFile(certFile, keyFile string) TransportOption {
	return func(opts *TransportOptions) error {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("failed to load certificate: %v", err)
		}

		if opts.TlsConfig == nil {
			opts.TlsConfig = &tls.Config{}
		}

		opts.TlsConfig.Certificates = []tls.Certificate{cert}
		return nil
	}
}

func WithSelfSignedCert() TransportOption {
	return func(opts *TransportOptions) error {
		if opts.TlsConfig == nil {
			opts.TlsConfig = &tls.Config{}
		}

		// Generate a self-signed certificate that expires in 30 days
		cert, err := mqc.GenerateCertificate(30 * 24 * time.Hour)
		if err != nil {
			return fmt.Errorf("failed to generate self-signed certificate: %v", err)
		}

		opts.TlsConfig.InsecureSkipVerify = true
		opts.TlsConfig.Certificates = []tls.Certificate{cert}

		return nil
	}
}

func WithOnConnect(f func(mqc.Transport)) TransportOption {
	return func(opts *TransportOptions) error {
		if f == nil {
			return fmt.Errorf("OnConnect function cannot be nil")
		}
		if opts.OnConnect != nil {
			return fmt.Errorf("OnConnect function is already set")
		}
		opts.OnConnect = f
		return nil
	}
}
