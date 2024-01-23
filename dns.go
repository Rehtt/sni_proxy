package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type dnsNode struct {
	A   []net.IP
	TTL time.Time
}

var dnsMap sync.Map

func QueryDns(domain string) ([]net.IP, error) {
	d, ok := dnsMap.Load(domain)

	cache, cacheOk := d.(*dnsNode)
	if ok && cacheOk && cache.TTL.After(time.Now()) {
		return cache.A, nil
	}
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	r, _, err := c.Exchange(m, "8.8.8.8:53")
	if err != nil {
		return nil, err
	}
	ips := r.Answer
	if len(ips) == 0 {
		return nil, fmt.Errorf("not find ip")
	}
	if cacheOk {
		cache.A = cache.A[:0]
	} else {
		cache = new(dnsNode)
		cache.A = make([]net.IP, 0, len(ips))
	}
	for _, v := range ips {
		cache.A = append(cache.A, v.(*dns.A).A)
	}
	cache.TTL = time.Now().Add(time.Duration(ips[0].Header().Ttl))
	dnsMap.Store(domain, cache)
	return cache.A, err
}
