FROM python:alpine

# upgrade packages, install ffmpeg, link lib64
RUN apk update && apk upgrade && \
    apk add ffmpeg && \
    mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

# put youtube-dl
ADD https://yt-dl.org/downloads/latest/youtube-dl  /usr/local/bin/youtube-dl

# put y2p
COPY y2p /usr/local/bin/y2p

# modify permission to allow execute
RUN adduser -u 14295 -DH -s /bin/sh y2p && \
    mkdir /data && chown -R y2p:y2p /data && \
    chmod a+rx /usr/local/bin/youtube-dl /usr/local/bin/y2p

# expose default port and config environment variable
EXPOSE 14295
ENV Y2P_CONFIG_PATH=/y2p-config.json

# execute
USER 14295:14295
CMD ["/usr/local/bin/y2p"]