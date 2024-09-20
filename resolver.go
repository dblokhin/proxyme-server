package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"

	lru "github.com/hashicorp/golang-lru/v2"
)

const maxCacheSize = 3000 // todo: parametrize

var defaultResolver = resolver{
	resolver: net.DefaultResolver,
	sg: &singleflight[string, []net.IP]{
		m: make(map[string]*singleflightResult[[]net.IP]),
	},
	cache: newSyncCache[string, []net.IP](maxCacheSize),
}

func newSyncCache[K comparable, V any](size int) *lru.Cache[K, V] {
	c, err := lru.New[K, V](size)
	if err != nil {
		panic(err)
	}

	return c
}

type resolver struct {
	resolver interface {
		LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
	}
	sg    *singleflight[string, []net.IP]
	cache *lru.Cache[string, []net.IP]
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
