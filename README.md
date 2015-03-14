# tv2go
Workalike to Sonarr ad Sickbeard - mostly to learn Go better

## What Works

You can add shows, get information from tvdb and tvrage, manually search for episodes and send them to a folder that SABnzbd is watching.  There is post processing functionality to import downloaded files but it's not well tested.

## What's needed

* More Providers, especially torrent providers. fanzub and nyaa torrents are at the top of the list
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

It currently works with Sabnzbd

Set the NZBBlackhole directory in config.json to where Sabnzbd will look for new .nzb files.  Create a subdirectory named tv2go
```
"NZBBlackhole": "/path/to/nzb/blackhole/tv2go"
mkdir -p /path/to/nzb/blackhole/tv2go
```

Compile the postprocess script and copy it to the Sabnzbd postprocess script directory

```
cd postprocess
go build -o sabToTv2go postprocess.go
cp sabToTv2go /path/to/postprocess/scripts
```

Create a new category in Sabnzbd named tv2go and set it to use sabToTv2go to postprocess the downloads.
