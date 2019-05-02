package threebot

import (
	"time"
	"encoding/json"
	"strings"
	"fmt"
	"net"

	"github.com/miekg/dns"

	"github.com/coredns/coredns/plugin"

	redisCon "github.com/garyburd/redigo/redis"
)

type Threebot struct {
	Next           plugin.Handler
	Pool           *redisCon.Pool
	ThreebotAddress   string
	ThreebotPassword  string
	connectTimeout int
	readTimeout    int
	keyPrefix      string
	keySuffix      string
	Ttl            uint32
	Zones          []string
	LastZoneUpdate time.Time
}

type Zone struct {
	Name      string
	Locations map[string]struct{}
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
	var (
		reply interface{}
		err error
		zones []string
	)

	conn := threebot.Pool.Get()
	if conn == nil {
		fmt.Println("error connecting to threebot")
		return
	}
	defer conn.Close()

	reply, err = conn.Do("KEYS", threebot.keyPrefix + "*" + threebot.keySuffix)
	if err != nil {
		return
	}
	zones, err = redisCon.Strings(reply, nil)
	for i, _ := range zones {
		zones[i] = strings.TrimPrefix(zones[i], threebot.keyPrefix)
		zones[i] = strings.TrimSuffix(zones[i], threebot.keySuffix)
	}
	threebot.LastZoneUpdate = time.Now()
	threebot.Zones = zones
}

func (threebot *Threebot) A(name string, z *Zone, record *Record) (answers, extras []dns.RR) {
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

func (threebot Threebot) AAAA(name string, z *Zone, record *Record) (answers, extras []dns.RR) {
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

func (threebot *Threebot) CNAME(name string, z *Zone, record *Record) (answers, extras []dns.RR) {
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

func (threebot *Threebot) hosts(name string, z *Zone) []dns.RR {
	var (
		record *Record
		answers []dns.RR
	)
	location := threebot.findLocation(name, z)
	if location == "" {
		return nil
	}
	record = threebot.get(location, z)
	a, _ := threebot.A(name, z, record)
	answers = append(answers, a...)
	aaaa, _ := threebot.AAAA(name, z, record)
	answers = append(answers, aaaa...)
	cname, _ := threebot.CNAME(name, z, record)
	answers = append(answers, cname...)
	return answers
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

func (threebot *Threebot) findLocation(query string, z *Zone) string {
	var (
		ok bool
		closestEncloser, sourceOfSynthesis string
	)

	// request for zone records
	if query == z.Name {
		return query
	}

	query = strings.TrimSuffix(query, "." + z.Name)

	if _, ok = z.Locations[query]; ok {
		return query
	}

	closestEncloser, sourceOfSynthesis, ok = splitQuery(query)
	for ok {
		ceExists := keyMatches(closestEncloser, z) || keyExists(closestEncloser, z)
		ssExists := keyExists(sourceOfSynthesis, z)
		if ceExists {
			if ssExists {
				return sourceOfSynthesis
			} else {
				return ""
			}
		} else {
			closestEncloser, sourceOfSynthesis, ok = splitQuery(closestEncloser)
		}
	}
	return ""
}

func (threebot *Threebot) get(key string, z *Zone) *Record {
	var (
		err error
		reply interface{}
		val string
	)
	conn := threebot.Pool.Get()
	if conn == nil {
		fmt.Println("error connecting to threebot")
		return nil
	}
	defer conn.Close()

	var label string
	if key == z.Name {
		label = "@"
	} else {
		label = key
	}

	reply, err = conn.Do("HGET", threebot.keyPrefix + z.Name + threebot.keySuffix, label)
	if err != nil {
		return nil
	}
	val, err = redisCon.String(reply, nil)
	if err != nil {
		return nil
	}
	r := new(Record)
	err = json.Unmarshal([]byte(val), r)
	if err != nil {
		fmt.Println("parse error : ", val, err)
		return nil
	}
	return r
}

func keyExists(key string, z *Zone) bool {
	_, ok := z.Locations[key]
	return ok
}

func keyMatches(key string, z *Zone) bool {
	for value := range z.Locations {
		if strings.HasSuffix(value, key) {
			return true
		}
	}
	return false
}

func splitQuery(query string) (string, string, bool) {
	if query == "" {
		return "", "", false
	}
	var (
		splits []string
		closestEncloser string
		sourceOfSynthesis string
	)
	splits = strings.SplitAfterN(query, ".", 2)
	if len(splits) == 2 {
		closestEncloser = splits[1]
		sourceOfSynthesis = "*." + closestEncloser
	} else {
		closestEncloser = ""
		sourceOfSynthesis = "*"
	}
	return closestEncloser, sourceOfSynthesis, true
}

func (threebot *Threebot) connect() {
	threebot.Pool = &redisCon.Pool{
		Dial: func () (redisCon.Conn, error) {
			opts := []redisCon.DialOption{}
			if threebot.ThreebotPassword != "" {
				opts = append(opts, redisCon.DialPassword(threebot.ThreebotPassword))
			}
			if threebot.connectTimeout != 0 {
				opts = append(opts, redisCon.DialConnectTimeout(time.Duration(threebot.connectTimeout)*time.Millisecond))
			}
			if threebot.readTimeout != 0 {
				opts = append(opts, redisCon.DialReadTimeout(time.Duration(threebot.readTimeout)*time.Millisecond))
			}

			return redisCon.Dial("tcp", threebot.ThreebotAddress, opts...)
		},
	}
}

func (threebot *Threebot) save(zone string, subdomain string, value string) error {
	var err error

	conn := threebot.Pool.Get()
	if conn == nil {
		fmt.Println("error connecting to threebot")
		return nil
	}
	defer conn.Close()

	_, err = conn.Do("HSET", threebot.keyPrefix + zone + threebot.keySuffix, subdomain, value)
	return err
}

func (threebot *Threebot) load(zone string) *Zone {
	var (
		reply interface{}
		err error
		vals []string
	)

	conn := threebot.Pool.Get()
	if conn == nil {
		fmt.Println("error connecting to threebot")
		return nil
	}
	defer conn.Close()

	reply, err = conn.Do("HKEYS", threebot.keyPrefix + zone + threebot.keySuffix)
	if err != nil {
		return nil
	}
	z := new(Zone)
	z.Name = zone
	vals, err = redisCon.Strings(reply, nil)
	if err != nil {
		return nil
	}
	z.Locations = make(map[string]struct{})
	for _, val := range vals {
		z.Locations[val] = struct{}{}
	}

	return z
}

func split255(s string) []string {
	if len(s) < 255 {
		return []string{s}
	}
	sx := []string{}
	p, i := 0, 255
	for {
		if i <= len(s) {
			sx = append(sx, s[p:i])
		} else {
			sx = append(sx, s[p:])
			break

		}
		p, i = p+255, i+255
	}

	return sx
}

const (
	defaultTtl = 360
	hostmaster = "hostmaster"
	zoneUpdateTime = 10*time.Minute
)