package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

const (
	dnsCacheSize = 3000 // todo: parametrize
	dnsCacheTTL  = 24 * time.Hour
)

var defaultResolver = resolver{
	resolver: net.DefaultResolver,
	sg: &singleflight[string, []net.IP]{
		m: make(map[string]*singleflightResult[[]net.IP]),
	},
	cache: expirable.NewLRU[string, []net.IP](dnsCacheSize, nil, dnsCacheTTL),
}

type resolver struct {
	resolver interface {
		LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
	}
	sg    *singleflight[string, []net.IP]
	cache *expirable.LRU[string, []net.IP]
}

// LookupIP resolves domain name
func (r *resolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	key := network + host

	if ips, ok := r.cache.Get(key); ok {
		return ips, nil
	}

	return r.sg.Do(key, func() ([]net.IP, error) {
		ips, err := r.resolver.LookupIP(ctx, network, host)
		if err != nil {
			return nil, err
		}

		if len(ips) == 0 {
			return nil, fmt.Errorf("failed to resolve %q", host)
		}

		r.cache.Add(key, ips)

		return ips, nil
	})
}

func defaultDomainResolver(ctx context.Context, domain []byte) (net.IP, error) {
	ips, err := defaultResolver.LookupIP(ctx, "ip", string(domain))
	if err != nil {
		return nil, err
	}

	// ipv4 priority
	for _, ip := range ips {
		if len(ip) == net.IPv4len {
			return ip, nil
		}
	}

	return ips[rand.Intn(len(ips))], nil // nolint
}
