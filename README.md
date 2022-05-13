# Cyber Stasis

![Logo](./assets/logo.png)

Cyber Stasis is an economic simulator in the form of a free fictional game based on global real-time demand and supply. The game tests the hypothesis of having a market system without a monetary one. There is no private property and a concept of wealth.
It's a pure market system focused on efficiency of distribution. Everything that we measure with money can be measured by a ratio between supply and demand. The goal of the system is to make sure that all needs are met to the best extent possible. There is a personal reputation index based on that which measures your contributions to society. The goal of the game is to become the most useful member of society. 

## How to Play

The game runs on the public IPFS network. In order to play it follow the steps below:

1. Install the official IPFS Desktop http://docs.ipfs.io/install/ipfs-desktop/
2. Install IPFS Companion http://docs.ipfs.io.ipns.localhost:8080/install/ipfs-companion/
3.  Clone https://github.com/stateless-minds/go-ipfs to your local machine, build it with `make build` and run it with the following command: `~/cmd/ipfs/ipfs daemon --enable-pubsub-experiment`
4.  Follow the instructions here to open your config file: https://github.com/ipfs/go-ipfs/blob/master/docs/config.md. Usually it's `~/.ipfs/config` on Linux. Add the <a href="https://ipfs.io/ipfs/QmQxd8Drsg5bvwqk3BbVmmNuouiJn9DdLY65717wWkLUxA">APP URL</a> to the `Access-Control-Allow-Origin` list
5.  Navigate to <a href="https://ipfs.io/ipfs/QmQxd8Drsg5bvwqk3BbVmmNuouiJn9DdLY65717wWkLUxA">APP URL</a> and let's simulate the future together!
6.  If you like the game consider pinning it to your local node so that you become a permanent host of it while you have IPFS daemon running
![SetPinning](./assets/set-pinning.png)
![PinToLocalNode](./assets/pin-to-local-node.png)

Please note the game has been developed on a WQHD resolution(2560x1440) and is currently not responsive or optimized for mobile devices. For best gaming experience if you play in FHD(1920x1080) please set your browser zoom settings to 150%.

## Guidelines

* **Economic simulator** - Cyber Stasis is an economic simulator in the form of a fictional game based on global real-time demand and supply.
* **Real-time demand/supply graph** - The graph reflects all demand and supply requests and is updated in real-time.
* **Supply can be sent only in response to an existing demand** - Send only goods and services you can provide in real life.
* **Keep it real** - Send requests for your real daily needs to make the whole simulation as accurate as possible.
* **Global events** - When the supply/demand ratio drops below certain thresholds global events are triggered and sent as notifications such as global shortage of water, food and housing.
* **Do what you do in real life** - Ask for things you need and supply things you provide.
* **Rankings** - Rankings reflect the level of contribution and usefulness of members to society. They take all factors into account and are calculated by a formula. The Reputation Index is the score in the game. Provide more than you consume and become the most valuable member of society!
* **No collection of user data** - Cyber Stasis does not collect any personal user data.
* **User generated content is fictional** - All user generated content is fictional and creators are responsibile for it.
* **If you like it support it** - This is an open source community project. Feel free to improve it or fork it and use it for your projects. Donations are welcome.

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
