package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/hobeone/tv2go/web"
)

// TODO:

func main() {
	defer glog.Flush()
	flag.Set("logtostderr", "true")
	web.StartServer()
	/*
		dbh := db.NewDBHandle("test.db", true, true)

		sl, err := tvdb.SearchSeries("Farscape", 10)
		if err != nil {
			panic(err)
		}
		spew.Dump(sl)

		series := sl.Series[0]

		newShow := db.Show{
			Name: series.SeriesName,
		}

		dbh.DB.Create(newShow)
	*/
}
