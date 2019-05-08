package threebot

import (
	"context"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// ServeDNS implements the plugin.Handler interface.
func (threebot *Threebot) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.Name()
	qtype := state.Type()

	zone := plugin.Zones(threebot.Zones).Matches(qname)
	location := threebot.findLocation(qname, zone)

	answers := make([]dns.RR, 0, 10)
	extras := make([]dns.RR, 0, 10)

	record, err := threebot.get(location)
	if err != nil {
		return threebot.errorResponse(state, zone, dns.RcodeBadName, nil)
	}

	switch qtype {
	case "A":
		answers, extras = threebot.A(qname, "", record)
	case "AAAA":
		answers, extras = threebot.AAAA(qname, "", record)
	case "CNAME":
		answers, extras = threebot.CNAME(qname, "", record)

	default:
		return threebot.errorResponse(state, zone, dns.RcodeNotImplemented, nil)
	}


	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = append(m.Answer, answers...)
	m.Extra = append(m.Extra, extras...)

	w.WriteMsg(m)
	return dns.RcodeSuccess, nil

}

// Name implements the Handler interface.
func (threebot *Threebot) Name() string { return "threebot" }

func (threebot *Threebot) errorResponse(state request.Request, zone string, rcode int, err error) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(state.Req, rcode)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true


	state.SizeAndDo(m)
	state.W.WriteMsg(m)
	return dns.RcodeSuccess, err
}