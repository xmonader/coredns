package threebot

import (
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/mholt/caddy"
)

var log = clog.NewWithPlugin("threebot")

func init() {
	caddy.RegisterPlugin("threebot", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

//func periodicHostsUpdate(h *Hosts) chan bool {
//	parseChan := make(chan bool)
//
//	if h.options.reload == durationOf0s {
//		return parseChan
//	}
//
//	go func() {
//		ticker := time.NewTicker(h.options.reload)
//		for {
//			select {
//			case <-parseChan:
//				return
//			case <-ticker.C:
//				h.readHosts()
//			}
//		}
//	}()
//	return parseChan
//}

func setup(c *caddy.Controller) error {
	c.Next() // 'threebot'
	if c.NextArg() {
		return plugin.Error("threebot", c.ArgErr())
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return Threebot{}
	})

	return nil
}