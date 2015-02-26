package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"
)

func TestNzbsOrg(t *testing.T) {
	//n := NewNzbsOrg("API_KEY", SetClient(&http.Client{}))
	//n.TvSearch("Archer (2009)", 5, 1)
	//
	content, err := ioutil.ReadFile("testdata/nzbs_org_archer.json")
	if err != nil {
		t.Fatalf("Error reading file: %s", err)
	}

	str := TvSearchResponse{}
	err = json.Unmarshal(content, &str)
	if err != nil {
		t.Fatalf("Error unmarshaling file: %s", err)
	}

	for _, item := range str.Channel.Items {
		fmt.Println(item.Title)
		fmt.Println(item.Link)
		fmt.Println(item.PubDate)
		pt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			t.Fatalf("Error unmarshaling file: %s", err)
		}
		fmt.Println(pt)

		fmt.Println(item.Category)
		fmt.Println(item.Enclosure.Attributes.Length)
	}
}
