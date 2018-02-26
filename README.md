# youtube2podcast
Make YouTube channels to become MP3 PodCast feeds

![youtube2podcast](https://github.com/xinsnake/youtube2podcast/raw/master/cover.png)

## Dependencies

- youtube-dl
- ffmpeg

## Usage

1. Edit the `y2p-config.sample.json` file
1. Set environment variable `Y2P_CONFIG_PATH` to point ot the configuration file
1. Run the application

or you can use docker

1. `docker run -d -v $(pwd)/y2p-config.sample.json:/y2p-config.json:ro -p 14295:14295 xinsnake/youtube2podcast`
1.1 Use `-v $(pwd)/assets:/assets:ro` if you want to change the template or styling
1.1 Use `-v $(pwd)/data:/data:raw` if you want to change data storage location

## Todo

- Removed feeds clean up
- Clean shutdown
- ~~Better home page~~
- ~~Directory clean up~~
- ~~Put the application in container~~