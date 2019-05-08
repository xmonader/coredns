package threebot

import (
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"strings"

	//"github.com/coredns/coredns/plugin/pkg/upstream"
	"github.com/mholt/caddy"
)

var log = clog.NewWithPlugin("threebot")

func init() {
	caddy.RegisterPlugin("threebot", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	r, err := threebotParse(c)
	if err != nil {
		return plugin.Error("threebot", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		r.Next = next
		return r
	})

	return nil
}

func threebotParse(c *caddy.Controller) (*Threebot, error) {
	threebot := &Threebot{Zones: []string{}}
	for c.Next() {
		if !c.NextArg() {       		   // expect at least one value
			return threebot, c.ArgErr()   // otherwise it's an error
		}
		// zone name
		value := c.Val()
		threebot.Zones = append(threebot.Zones, value)
		for i, str := range threebot.Zones {
			threebot.Zones[i] = plugin.Host(str).Normalize()
		}

		for c.NextBlock() {
			switch c.Val() {
			case "explorer":
				if !c.NextArg(){
					return threebot, c.ArgErr()
				}
				threebot.Explorers = append(threebot.Explorers, strings.TrimRight(c.Val(), "/"))

			default:
				return threebot, c.ArgErr() //invalid argument.
			}
		}
	}
	return threebot, nil
}