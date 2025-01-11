package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/sync/singleflight"
)

const (
	dnsCacheSize = 3000 // todo: parametrize
	dnsCacheTTL  = 24 * time.Hour
)

var defaultResolver = resolver{
	resolver: net.DefaultResolver,
	sg:       new(singleflight.Group),
	cache:    expirable.NewLRU[string, []net.IP](dnsCacheSize, nil, dnsCacheTTL),
}

type resolver struct {
	resolver interface {
		LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
	}
	sg    *singleflight.Group
	cache *expirable.LRU[string, []net.IP]
}

// LookupIP resolves domain name
func (r *resolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	key := network + host

	if ips, ok := r.cache.Get(key); ok {
		return ips, nil
	}

	res, err, _ := r.sg.Do(key, func() (interface{}, error) {
		ips, err := r.resolver.LookupIP(ctx, network, host)
		if err != nil {
			return nil, err
		}

		if len(ips) == 0 {
			return nil, fmt.Errorf("failed to resolve %q", host)
		}

		r.cache.Add(key, ips)
		r.sg.Forget(key)

		return ips, nil
	})

	if res != nil {
		return res.([]net.IP), err
	}

	return nil, err
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
