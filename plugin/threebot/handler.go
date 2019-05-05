package threebot

import (
	"fmt"
	// "fmt"
	//"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
	"context"
	"github.com/coredns/coredns/request"
)

// ServeDNS implements the plugin.Handler interface.
func (threebot *Threebot) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	fmt.Println("serveDNS")
	state := request.Request{W: w, Req: r}
	zone := "grid.tf."
	qname := state.Name()
	qtype := state.Type()

	fmt.Println("name : ", qname)
	fmt.Println("type : ", qtype)

	//
	//zone = plugin.Zones(threebot.Zones).Matches(qname)
	//fmt.Println("zone : ", zone)
	//if zone == "" {
	//	return plugin.NextOrFailure(qname, threebot.Next, ctx, w, r)
	//}


	location := threebot.findLocation(qname, zone)
	fmt.Println("LOCATION SERVDNS: ", location)
	//if len(location) == 0 { // empty, no results
	//	return threebot.errorResponse(state, zone, dns.RcodeNameError, nil)
	//}
	// fmt.Println("location : ", location)

	answers := make([]dns.RR, 0, 10)
	extras := make([]dns.RR, 0, 10)

	record := threebot.get(location)
	fmt.Println("Record: ", record)
	fmt.Println("RECORD: ",  record.A )

	switch qtype {
	case "A":
		answers, extras = threebot.A(qname, "", record)
	case "AAAA":
		answers, extras = threebot.AAAA(qname, "", record)
	case "CNAME":
		answers, extras = threebot.CNAME(qname, "", record)

	//case "NS":
	//	answers, extras = threebot.NS(qname, z, record)
	//case "MX":
	//	answers, extras = threebot.MX(qname, z, record)
	//case "SRV":
	//	answers, extras = threebot.SRV(qname, z, record)
	//case "SOA":
	//	answers, extras = threebot.SOA(qname, z, record)
	//case "CAA":
	//	answers, extras = threebot.CAA(qname, z, record)
	default:
		return threebot.errorResponse(state, zone, dns.RcodeNotImplemented, nil)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true


	m.Answer = append(m.Answer, answers...)
	m.Extra = append(m.Extra, extras...)

	//fmt.Printf("%v %v %v \n", m, m.Answer, m.Extra)

	state.SizeAndDo(m)
	m = state.Scrub(m)
	fmt.Println("WRITING MSG NOW", m)
	w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (threebot *Threebot) Name() string { return "threebot" }

func (threebot *Threebot) errorResponse(state request.Request, zone string, rcode int, err error) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(state.Req, rcode)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	// m.Ns, _ = threebot.SOA(state.Name(), zone, nil)

	state.SizeAndDo(m)
	state.W.WriteMsg(m)
	// Return success as the rcode to signal we have written to the client.
	return dns.RcodeSuccess, err
}