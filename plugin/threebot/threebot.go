package threebot

import (
	"encoding/json"
	"fmt"
	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
	"net"
	"net/http"
	"strings"
)

type Threebot struct {
	Next           plugin.Handler
	Ttl            uint32
	Zones          []string
	Explorers	   []string
}

type Zone struct {
	Name      string
	Locations map[string]Record
}

type Record struct {
	A     []A_Record `json:"a,omitempty"`
	AAAA  []AAAA_Record `json:"aaaa,omitempty"`
	CNAME []CNAME_Record `json:"cname,omitempty"`

}

type A_Record struct {
	Ttl uint32 `json:"ttl,omitempty"`
	Ip  net.IP `json:"ip"`
}

type AAAA_Record struct {
	Ttl uint32 `json:"ttl,omitempty"`
	Ip  net.IP `json:"ip"`
}

type CNAME_Record struct {
	Ttl  uint32 `json:"ttl,omitempty"`
	Host string `json:"host"`
}

func (threebot *Threebot) A(name, z string,  record *Record) (answers, extras []dns.RR) {
	for _, a := range record.A {
		if a.Ip == nil {
			continue
		}
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeA,
			Class: dns.ClassINET, Ttl: threebot.minTtl(a.Ttl)}
		r.A = a.Ip
		answers = append(answers, r)
	}
	return
}

func (threebot Threebot) AAAA(name, z string,  record *Record) (answers, extras []dns.RR) {
	for _, aaaa := range record.AAAA {
		if aaaa.Ip == nil {
			continue
		}
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA,
			Class: dns.ClassINET, Ttl: threebot.minTtl(aaaa.Ttl)}
		r.AAAA = aaaa.Ip
		answers = append(answers, r)
	}
	return
}

func (threebot *Threebot) CNAME(name, z string,  record *Record) (answers, extras []dns.RR) {
	for _, cname := range record.CNAME {
		if len(cname.Host) == 0 {
			continue
		}
		r := new(dns.CNAME)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeCNAME,
			Class: dns.ClassINET, Ttl: threebot.minTtl(cname.Ttl)}
		r.Target = dns.Fqdn(cname.Host)
		answers = append(answers, r)
	}
	return
}

func (threebot *Threebot) minTtl(ttl uint32) uint32 {
	if threebot.Ttl == 0 && ttl == 0 {
		return defaultTtl
	}
	if threebot.Ttl == 0 {
		return ttl
	}
	if ttl == 0 {
		return threebot.Ttl
	}
	if threebot.Ttl < ttl {
		return threebot.Ttl
	}
	return  ttl
}

func (threebot *Threebot) findLocation(query, zoneName string) string {
	// request for zone records
	if query == zoneName {
		return query
	}

	query = strings.TrimSuffix(query, "." + zoneName)
	if strings.Count(query, ".") == 1{
		return query
	}
	return ""
}

func (threebot *Threebot) get(key string) (*Record, error) {
	/*
	https://explorer.testnet.threefoldtoken.com/explorer/whois/3bot/zaibon.tf3bot
	{"record":{"id":1,"addresses":["3bot.zaibon.be"],"names":["zaibon.tf3bot"],"publickey":"ed25519:ea07bcf776736672370866151fc6850347eae36dda2a0653113102ea84d8ac1c","expiration":1559052900}}
	*/
	type ThreeBotRecord struct {
		Addresses []string `json:"addresses"`
		Names     []string `json:"names"`
	}
	type WhoIsResponse struct{
		ThreeBotRecord `json:"record"`
	}
	// whoever responds is enough
	rec := new(Record)
	rec.A = []A_Record{}
	rec.AAAA = []AAAA_Record{}
	for _, explorer := range threebot.Explorers {
		whoisUrl := explorer+"/explorer/whois/3bot/"+key
		resp, error := http.Get(whoisUrl)
		defer resp.Body.Close()

		if error != nil {
			return nil, error
		}
		if resp.StatusCode==200{
			whoisResp := new(WhoIsResponse)
			if err := json.NewDecoder(resp.Body).Decode(whoisResp); err != nil {
				continue
			}
			// TODO: handle multiple records and agree on standard return of IPv4 for locations where 3bots are running on.
			for _, addr := range(whoisResp.Addresses) {

				theIp := net.ParseIP(addr)
				if theIp != nil {
					if  theIp.To4() != nil {
						rec.A = append(rec.A, A_Record{Ip: theIp, Ttl: 300})
						continue
					}
					if  theIp.To16() != nil {
						rec.AAAA = append(rec.AAAA, AAAA_Record{Ip: theIp})
						continue
					}
				}
				// TODO: hostnames.
			}
		}
	}
	if len(rec.A) + len(rec.AAAA) > 0 {
		return rec, nil
	}

	return nil, fmt.Errorf("couldn't get record for 3bot with key %s ", key)
}

const (
	defaultTtl = 360
)