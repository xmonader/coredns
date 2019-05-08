package threebot

import (
	"github.com/mholt/caddy"
	"testing"
)

const (
	simpleCorefile = `
    threebot grid.tf. {
    	explorer https://explorer.testnet.threefoldtoken.com
    }

`
	zoneName = "grid.tf."
	explorer1Url = "https://explorer.testnet.threefoldtoken.com"
	explorer2Url = "https://explorer1.testnet.threefoldtoken.com"

	multipleExplorerCorefile = `
    threebot grid.tf. {
    	explorer https://explorer.testnet.threefoldtoken.com
		explorer https://explorer1.testnet.threefoldtoken.com

    }
`
	trailingSlashInCorefileExplorers = `
    threebot grid.tf. {
    	explorer https://explorer.testnet.threefoldtoken.com/
		explorer https://explorer1.testnet.threefoldtoken.com/

    }
`
)


func TestParseThreebotConfigWithSingleExplorer(t *testing.T){
	c := caddy.NewTestController("dns", simpleCorefile)
	threebot, error := threebotParse(c)
	if error != nil {
		t.Errorf("Couldn't parse Corefile %s", error)
	}
	if len(threebot.Zones) != 1 {
		t.Errorf("Expected to have 1 zone `%s`", zoneName)
	}
	if threebot.Zones[0] != zoneName {
		t.Errorf("Zone expected to be %s but got %s", zoneName, threebot.Zones[0])
	}
	if len(threebot.Explorers) != 1 {
		t.Errorf("Expected to have 1 explorer.")
	}
	if threebot.Explorers[0] != explorer1Url {
		t.Errorf("Expected the explorer to be %s but got %s", explorer1Url, threebot.Explorers[0])
	}
}
func TestParseThreebotConfigWithMultipleExplorers(t *testing.T){
	c := caddy.NewTestController("dns", multipleExplorerCorefile)
	threebot, error := threebotParse(c)
	if error != nil {
		t.Errorf("Couldn't parse Corefile %s", error)
	}
	if len(threebot.Zones) != 1 {
		t.Errorf("Expected to have 1 zone `%s`", zoneName)
	}
	if threebot.Zones[0] != zoneName {
		t.Errorf("Zone expected to be %s but got %s", zoneName, threebot.Zones[0])
	}
	if len(threebot.Explorers) != 2 {
		t.Errorf("Expected to have 2 explorer.")
	}
	if threebot.Explorers[0] != explorer1Url {
		t.Errorf("Expected Explorers[0] to be %s but got %s", explorer1Url, threebot.Explorers[0])
	}
	if threebot.Explorers[1] != explorer2Url {
		t.Errorf("Expected Explorers[1] to be %s but got %s", explorer1Url, threebot.Explorers[1])

	}
}

func TestParseThreebotConfigMultipleExplorersWithTrailingSlash(t *testing.T) {
	c := caddy.NewTestController("dns", trailingSlashInCorefileExplorers)
	threebot, error := threebotParse(c)
	if error != nil {
		t.Errorf("Couldn't parse Corefile %s", error)
	}
	if len(threebot.Zones) != 1 {
		t.Errorf("Expected to have 1 zone `%s`", zoneName)
	}
	if threebot.Zones[0] != zoneName {
		t.Errorf("Zone expected to be %s but got %s", zoneName, threebot.Zones[0])
	}
	if len(threebot.Explorers) != 2 {
		t.Errorf("Expected to have 2 explorer.")
	}
	if threebot.Explorers[0] != explorer1Url {
		t.Errorf("Expected Explorers[0] to be %s but got %s", explorer1Url, threebot.Explorers[0])
	}
	if threebot.Explorers[1] != explorer2Url {
		t.Errorf("Expected Explorers[1] to be %s but got %s", explorer1Url, threebot.Explorers[1])

	}

}