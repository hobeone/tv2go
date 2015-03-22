package daemon

import (
	"testing"

	"github.com/hobeone/tv2go/config"
	"github.com/hobeone/tv2go/db"
	"github.com/hobeone/tv2go/providers"
	. "github.com/onsi/gomega"
)

func TestProcessProviderResult(t *testing.T) {
	RegisterTestingT(t)

	cfg := config.NewTestConfig()
	//cfg.DB.Verbose = true
	cfg.Providers = append(cfg.Providers, config.ProviderConfig{Name: "nzbsOrg", API: "123"})
	d := NewDaemon(cfg)

	db.LoadFixtures(t, d.DBH)

	d.DBH.SaveNameExceptions("xem_tvdb", []*db.NameException{
		&db.NameException{
			Source:    "xem_tvdb",
			Indexer:   "tvdb",
			IndexerID: 2,
			Name:      "Yowamushi Pedal - Grande Road",
			Season:    2,
		},
	})

	pr := providers.ProviderResult{
		Name:         "[HorribleSubs] Yowamushi Pedal - Grande Road - 1 [720p].mkv",
		Anime:        true,
		ProviderName: "nyaaTorrents",
	}
	err := d.ProcessProviderResult(pr)
	Expect(err).To(MatchError("Couldn't download : Get : unsupported protocol scheme \"\""))
}
