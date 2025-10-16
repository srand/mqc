package transport

import "time"

type DialOptions struct {
	Addrs []string

	// Timeout for the dial operation
	ConnectTimeout time.Duration // in seconds

	// Timeout for individual RPC calls
	CallTimeout time.Duration // in seconds

	// Underlying protocol to use (e.g. "tcp", "unix").
	// Varies depending on the transport implementation.
	Protocol string
}

type DialOption func(*DialOptions)

func WithAddress(addr string) DialOption {
	return func(opts *DialOptions) {
		opts.Addrs = append(opts.Addrs, addr)
	}
}

func WithConnectTimeout(d time.Duration) DialOption {
	return func(opts *DialOptions) {
		opts.ConnectTimeout = d
	}
}

func WithCallTimeout(d time.Duration) DialOption {
	return func(opts *DialOptions) {
		opts.CallTimeout = d
	}
}

func WithProtocol(protocol string) DialOption {
	return func(opts *DialOptions) {
		opts.Protocol = protocol
	}
}
