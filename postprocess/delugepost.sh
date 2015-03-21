#!/bin/bash -ex

# Deluge provides arguments in a different order than sabnzbd, this is a quick
# and dirty wrapper to arrange things the way the post process script expects.

pushd $(dirname $0)

torrentid=$1
torrentname=$2
torrentpath=$3
echo "Torrent Details: " "$torrentpath/$torrentname"
go run postprocess.go "$torrentpath/$torrentname"
popd
