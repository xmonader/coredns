package threebot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"net"
	"github.com/miekg/dns"
	"github.com/coredns/coredns/plugin"
	deb "runtime/debug"

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
	fmt.Println("LEN RECORD A: ", len(record.A))
	for _, a := range record.A {
		if a.Ip == nil {
			continue
		}
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: name, Rrtype: dns.TypeA,
			Class: dns.ClassINET, Ttl: threebot.minTtl(a.Ttl)}
		fmt.Println("IP IN A FUNCT: ", string(a.Ip)	)
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
//
//func (threebot *Threebot) hosts(name, z string) []dns.RR {
//	fmt.Println("calling hosts..")
//	var (
//		record *Record
//		answers []dns.RR
//	)
//	location := threebot.findLocation(name, z)
//	if location == "" {
//		return nil
//	}
//	record = threebot.get(location)
//	a, _ := threebot.A(name, z, record)
//	answers = append(answers, a...)
//	aaaa, _ := threebot.AAAA(name, z, record)
//	answers = append(answers, aaaa...)
//	cname, _ := threebot.CNAME(name, z, record)
//	answers = append(answers, cname...)
//	return answers
//}

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
	fmt.Println("findLocation :" , query, " " , zoneName)
	if query == zoneName {
		fmt.Println("findlocation returning zoneName: ")
		return query
	}

	query = strings.TrimSuffix(query, "." + zoneName)
	fmt.Println("QUERY NOW: ", query)
	if strings.Count(query, ".") == 1{
		fmt.Println("returning query: ", query)
		return query
	}
	return ""
}

func (threebot *Threebot) get(key string) *Record {
	deb.PrintStack()

	fmt.Println("threebot get: ", key)
	// check errors.
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
		fmt.Println("RESP: ",string(body))

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
	reply := Record{
		A: []A_Record{
			{Ip: []byte("192.52.12.4"), Ttl:500},
		},
	}
	fmt.Println("returning reply: %v ", reply)
	return &reply
}

const (
	defaultTtl = 360
	hostmaster = "hostmaster"
	zoneUpdateTime = 10*time.Minute
)