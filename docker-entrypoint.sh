#!/bin/bash

while true
do
    if [ $STREAMING -eq 1 ]
    then
        tmpdir=`mktemp -d`
        ./run-stream-server.sh $tmpdir &
        ffmpeg_pid=$!

        sleep 5
        echo "Stream started on rtmp://0.0.0.0:1935/live/app"
        while [ ! -f $tmpdir/playlist.mpd ]
        do
            sleep 1
        done

        sleep 2
        ./server -dash $tmpdir/playlist.mpd -streaming -tls-cert /certs/cert.pem -tls-key /certs/key.pem &
        server_pid=$!

        wait $ffmpeg_pid
        kill $server_pid
        rm -rf $tmpdir
    else
        ./server -dash $DASH_PATH -tls-cert /certs/cert.pem -tls-key /certs/key.pem
    fi

done