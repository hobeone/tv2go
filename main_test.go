package main

/*
func TestGetNewItems(t *testing.T) {
	//flag.Set("logtostderr", "true")
	body, err := ioutil.ReadFile("providers/testdata/nzbs_org_feed_single.rss")
	if err != nil {
		t.Fatalf("Error reading test file %s", err)
	}
	flag.Set("logtostderr", "true")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(200)
		w.Write(body)
		fmt.Println(r.URL.String())
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	httpClient := &http.Client{Transport: transport}

	n := providers.NewNzbsOrg("API_KEY")
	n.Client = httpClient
	n.URL = server.URL

	provReg := providers.ProviderRegistry{
		// get key from cfg
		"nzbsOrg":      n,
		"nyaaTorrents": providers.NewNyaaTorrents(),
	}

	dbshow := &db.Show{
		Name:      "New Girl",
		IndexerID: 248682,
		Indexer:   "tvdb",
		Episodes: []db.Episode{
			{
				Name:    "Test",
				Season:  4,
				Episode: 10,
				Status:  types.WANTED,
			},
		},
	}

	b, _ := storage.NewBroker("/tmp")

	cfg := config.NewConfig()
	cfg.DB.Type = "memory"
	d := NewDaemon(cfg)
	err = d.DBH.AddShow(dbshow)
	if err != nil {
		t.Fatalf("Error adding test show: %s", err)
	}
	d.ProviderPoller(&provReg, b)
}
*/
