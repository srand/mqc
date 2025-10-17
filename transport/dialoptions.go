package transport

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/srand/mqc"
)

type DialOptions struct {
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
}

type DialOption func(*DialOptions) error

func WithAddress(addr string) DialOption {
	return func(opts *DialOptions) error {
		opts.Addrs = append(opts.Addrs, addr)
		return nil
	}
}

func WithConnectTimeout(d time.Duration) DialOption {
	return func(opts *DialOptions) error {
		opts.ConnectTimeout = d
		return nil
	}
}

func WithCallTimeout(d time.Duration) DialOption {
	return func(opts *DialOptions) error {
		opts.CallTimeout = d
		return nil
	}
}

func WithProtocol(protocol string) DialOption {
	return func(opts *DialOptions) error {
		opts.Protocol = protocol
		return nil
	}
}

func WithTLSConfig(tlsConfig *tls.Config) DialOption {
	return func(opts *DialOptions) error {
		opts.TlsConfig = tlsConfig
		return nil
	}
}

func WithCertificateFile(certFile, keyFile string) DialOption {
	return func(opts *DialOptions) error {
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

func WithSelfSignedCert() DialOption {
	return func(opts *DialOptions) error {
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
