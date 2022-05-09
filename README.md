# Cyber Stasis

![Logo](./assets/logo.png)

Cyber Stasis is an economic simulator in the form of a free fictional game based on global real-time demand and supply.

## How to Play

The game runs on the public IPFS network. In order to play it follow the steps below:

1. Install the official IPFS Desktop http://docs.ipfs.io/install/ipfs-desktop/
2. Install IPFS Companion http://docs.ipfs.io.ipns.localhost:8080/install/ipfs-companion/
3.  Clone https://github.com/stateless-minds/go-ipfs to your local machine, build it with `make build` and run it with the following command: `~/cmd/ipfs/ipfs daemon --enable-pubsub-experiment`
4.  Follow the instructions here to open your config file: https://github.com/ipfs/go-ipfs/blob/master/docs/config.md. Usually it's `~/.ipfs/config` on Linux. Add the <a href="https://ipfs.io/ipfs/QmQxd8Drsg5bvwqk3BbVmmNuouiJn9DdLY65717wWkLUxA">APP URL</a> to the `Access-Control-Allow-Origin` list
5.  Navigate to <a href="https://ipfs.io/ipfs/QmQxd8Drsg5bvwqk3BbVmmNuouiJn9DdLY65717wWkLUxA">APP URL</a> and let's simulate the future together!

Please note the game has been developed on a WQHD resolution(2560x1440) and is currently not responsive or optimized for mobile devices. For best gaming experience if you play in FHD(1920x1080) please set your browser zoom settings to 150%.

## Screenshots

<a display="inline" href="./assets/home.png?raw=true">
<img src="./assets/home.png" width="45%" alt="Screenshot of the homepage" title="Screenshot of the homepage">
</a>

<a display="inline" href="./assets/ranks.png?raw=true">
<img src="./assets/ranks.png" width="45%" alt="Screenshot of the ranks screen" title="Screenshot of the ranks screen">
</a>

<a display="inline" href="./assets/story.png?raw=true">
<img src="./assets/story.png" width="45%" alt="Screenshot of the story" title="Screenshot of the story">
</a>

<a display="inline" href="./assets/mission.png?raw=true">
<img src="./assets/mission.png" width="45%" alt="Screenshot of the mission" title="Screenshot of the mission">
</a>

## Roadmap
1. Make it responsive - Not started
2. Make it mobile friendly - Not started. It will require a very different approach with a different client: https://berty.tech/docs/gomobile-ipfs/

## Acknowledgments

1. <a href="https://go-app.dev/">go-app</a>
2. <a href="https://ipfs.io/">IPFS</a>
3. <a href="https://berty.tech/">Berty</a>
4. All the rest of the authors who worked on the dependencies used! Thanks a lot!

## Contributing

<a href="https://github.com/stateless-minds/cyber-stasis/issues">Open an issue</a>

## License

Stateless Minds (c) 2022 and contributors

MIT License
