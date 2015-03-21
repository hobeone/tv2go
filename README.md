# tv2go
Workalike to Sonarr ad Sickbeard - mostly to learn Go better

https://godoc.org/github.com/hobeone/tv2go

[![wercker status](https://app.wercker.com/status/5a29ff8c01f89263d1178da25ffdadce/m "wercker status")](https://app.wercker.com/project/bykey/5a29ff8c01f89263d1178da25ffdadce)

## What Works

You can add shows, get information from tvdb and tvrage, manually search for episodes and send them to a folder that SABnzbd or a torrent downloader is watching.  There is post processing functionality to import downloaded files but is mostly tested with SABnzbd. See https://github.com/hobeone/tv2go/tree/master/postprocess for more info.

## What's needed

* More Providers, especially torrent providers.
* Better support for non standard show formats - daily, sports etc.
* A Javascript UI written by somebody that knows what they're doing
* More tests
* Better Anime Support

## Install instructions:

```
go get -t github.com/hobeone/tv2go

cd $GOPATH/src/github.com/hobeone/tv2go/
cp config_example.json config.json

edit config.json for file paths etc

cd webapp

npm install
```

```
To run during development:
go get -u -v github.com/codegangsta/gin

$GOPATH/bin/gin -i -p 9001 -a 9000 r -- -logtostderr=true -config_file config.json

Point your browser at localhost:9001/a/tv2go.html

Add shows and click around
```

It currently works with Sabnzbd or a torrent downloader that can watch a director for new torrents.

Set the NZBBlackhole directory in config.json to where Sabnzbd will look for new .nzb files.  Create a subdirectory named tv2go
```
"NZBBlackhole": "/path/to/nzb/blackhole/tv2go"
"TorrentBlackhole": "/path/to/torrent/blackhole"
mkdir -p /path/to/nzb/blackhole/tv2go
mkdir -p /path/to/torrent/blackhole
```


Compile the postprocess script and copy it and config.yaml to the Sabnzbd postprocess script directory

```
cd postprocess
go build -o sabToTv2go postprocess.go
cp sabToTv2go config.yaml /path/to/postprocess/scripts
```

Create a new category in Sabnzbd named tv2go and set it to use sabToTv2go to postprocess the downloads.

deluge can run scripts on torrent completion with the 'execute' plugin.  Set that up to run the delugepost.sh shell script and it should send downloaded files to tv2go for processing.

***
System Walkthrough

I primarily wrote this to learn [Go](http://golang.org) better.  It's been a very interesting process especially learning how to structure code given the lack of the type of OO code structure that I'm used to (Python and Ruby).


###Major Parts
- Name Parsing
- Indexers
- Providers
- Daemon
- REST API
- Web UI

###Name Parsing

Reliably Parsing strings into their various parts is the heart of a program like tv2go.  This is handled by the NameParser package of tv2go.  It is essentially a series of regular expressions and application of best guesses to try and extract the most information from a given string.

###Indexers

In tv2go indexers are services like http://thetvdb.com and http://www.tvrage.com.  They provide the canonical list of available shows and episodes.  These are consulted when adding a new show for tv2go to tack and on a regular background timer to update information about those shows.

In tv2go indexer libraries are written to convert information from the given indexer format to a canonical internal format that can be used by the rest of the system.

###Providers

Providers are sites that list shows available for download.  These are either [NZB](https://en.wikipedia.org/wiki/NZB) or [Torrent](https://en.wikipedia.org/wiki/BitTorrent) based sites.  The Provider interface in tv2go allows for search of a particular show/season/episode as well as polling the Provider every N minutes for new releases.

###Daemon

The Daemon is the part that ties everything together and sets up the background processing to watch for new releases from providers, update show information etc.

###REST API
All of the UI for tv2go is implemented over this API.  The API surface is pretty small, see configGinEngine in https://github.com/hobeone/tv2go/blob/master/web/web.go for everything.

###Web UI
This is an abomination for before Dog and Man.  Switching between Javascript and Go is incredibly jarring and I'm no good at Javascript anyway.
