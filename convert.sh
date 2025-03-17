ffmpeg -i ./playground/trim_33CDFC8B_3433_4A70_8141_EBAE39AC85AF.mov \
-c:v libx264 -preset slow -profile:v main \
-b:v 5000k -maxrate 5000k -bufsize 5000k \
-c:a aac -b:a 128k -ar 44100 -ac 2 \
-f mpegts -vf "scale=1280x720" -r 30 \
-flags +global_header -fflags +genpts \
./playground/video3.ts