package threebot

import (
	"encoding/json"
	"fmt"
	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

type Threebot struct {
	Next           plugin.Handler
	connectTimeout int
	readTimeout    int
	Ttl            uint32
	Zones          []string
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


func (threebot *Threebot) LoadZones() {
	zones := []string{"grid.tf."}

	threebot.Zones = zones
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

func (threebot *Threebot) get(key string) *Record {

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

	resp, err := http.Get("https://explorer.testnet.threefoldtoken.com/explorer/whois/3bot/"+key)
	if resp.StatusCode==200{
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err.Error())
		}
		whoisResp := new(WhoIsResponse)
		err = json.Unmarshal([]byte(body),&whoisResp)
		// check err
		fmt.Println(err)
	}
	if err!=nil {
		// todo
	}

	// TODO: handle multiple records and agree on standard return of IPv4 for locations where 3bots are running on.
	rec := new(Record)

	rec.A = []A_Record{
		{Ip: net.ParseIP("8.8.8.8"), Ttl:300},
	}
	return rec
}

const (
	defaultTtl = 360
)