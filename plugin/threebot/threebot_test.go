package threebot

import (
	"testing"
)


func TestParsingThreebotRecordsResponse(t *testing.T){
	/*
	{"record":{"id":38,
		"addresses": ["google.com","8.8.8.8","192.168.12.42","2001:db8:85a3::8a2e:370:7334"],
		"names":["codepaste.thabeta"],
		"publickey":"ed25519:47ae06c4457f8fc1ec9ecc944fc05459d320670575a95b517465b6c332a7f2d2","expiration":1559901840}}
	*/
	resp := &WhoIsResponse{
		ThreeBotRecord{
			Names:[]string{"codepaste.thabeta"},
			Addresses: []string{"google.com","8.8.8.8","192.168.12.42","2001:db8:85a3::8a2e:370:7334"},
		},
	}
	rec, error := recordsFromWhoIsResponse(resp)
	if error != nil {
		t.Errorf("%v whoIsResponse should be parsed correctly but got error %s", resp, error)

	}
	if len(rec.A) != 2{
		t.Errorf("%v record should have 2 A records but got %d",resp, len(rec.A))
	}

	if len(rec.AAAA) != 1{
		t.Errorf("%v record should have 1 AAAA record but got %d", resp, len(rec.AAAA))
	}

}