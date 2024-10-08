#!/bin/bash

while true
do
    tmpdir=`mktemp -d`
    ./media/run-stream-server.sh $tmpdir &
    ffmpeg_pid=$!

    sleep 5
    echo "Stream started on rtmp://0.0.0.0:1935/live/app"
    while [ ! -f $tmpdir/playlist.mpd ]
    do
        sleep 1
    done

    sleep 2
    ./server/server -dash $tmpdir/playlist.mpd -streaming $@ &
    server_pid=$!

    wait $ffmpeg_pid
    kill $server_pid
    rm -rf $tmpdir
done