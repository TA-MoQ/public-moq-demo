#!/bin/bash
input_file="source.mp4"
segment_duration=2
chunk_duration=1
fps=25
dir=${1:-'.'}

pkill ffmpeg
rm -f "$dir/*.m4s"
rm -f "$dir/*.tmp"
rm "$dir/playlist.mpd"


ffmpeg -listen 1 -i rtmp://0.0.0.0:1935/live/app \
    -f dash -ldash 1 \
    -c:v libx264 \
    -filter:v fps=$fps \
    -preset veryfast -tune zerolatency \
    -c:a aac \
    -b:a 128k -ac 2 -ar 44100 \
    -map v:0 -s:v:1 1080x720 -b:v:1 2.6M \
    -map v:0 -s:v:3 640x360  -b:v:3 365k \
    -map 0:a \
    -force_key_frames "expr:gte(t,n_forced*2)" \
    -sc_threshold 0 \
    -streaming 1 \
    -use_timeline 0 \
    -seg_duration $segment_duration -frag_duration $chunk_duration \
    -frag_type duration \
    "$dir/playlist.mpd"

