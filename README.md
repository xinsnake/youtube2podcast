# youtube2podcast
Make YouTube channels to become MP3 PodCast feeds

## Dependencies

- youtube-dl
- ffmpeg

## Usage

1. Edit the `y2p-config.sample.json` file
1. Set environment variable `Y2P_CONFIG_PATH` to point ot the configuration file
1. Run the application

or you can use docker

1. `docker run -d -v $(pwd)/y2p-config.sample.json:/y2p-config.json:ro -p 14295:14295 xinsnake/youtube2podcast`

## Todo

- --Directory clean up--
- Clean shutdown
- Put the application in container