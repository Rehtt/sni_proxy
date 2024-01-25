package main

import (
	"fmt"
	"net"
	"time"

	"github.com/Rehtt/Kit/cache"
	"github.com/miekg/dns"
)

var dnsCache *cache.Cache

func init() {
	dnsCache = cache.NewCache(60*time.Second, 60*time.Second)
}

func QueryDns(domain string) ([]net.IP, error) {
	d, ok := dnsCache.Get(domain)

	cache, cacheOk := d.([]net.IP)
	if cacheOk && ok {
		return cache, nil
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

	for _, v := range ips {
		cache = append(cache, v.(*dns.A).A)
	}
	dnsCache.Set(domain, cache, time.Duration(ips[0].Header().Ttl)*time.Second)
	return cache, err
}
