#!/bin/bash

cd ./media
./run-stream-server.sh &
cd ..

sleep 5
echo "Stream started on rtmp://0.0.0.0:1935/live/app"
read -p "Start streaming and press enter..."
./server/server -dash ./media/playlist.mpd -streaming $@