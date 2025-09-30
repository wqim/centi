# Centi

## What?
Centi is a doomsday hidden network which works over the publicy known platforms like HuggingFace, GitHub, etc.
Potentialy it can support any type of API, the only thing you need to change in case of specified service blocking
is just include another module for Centi network and recompile the program.
Centi also can run over TLS. Bluetooth support development is in progress.

## Purpose of the project
The purpose of the project is to create an absolute anonymity in the internet even in heavily censured countries.

>[!WARNING]
>This is an experimental project. Do not rely on Centi for important actions.

## Current state & development
The project is in active development. It has basic send/receive functionality but only with Gitea and TLS modules.

## How does this even work?
The Centi treats account like hidden network treats an IP address: it is used only to send and receive data in encrypted form.
It does not include actual address inside network packets, even your public key can be unknown to anyone when you connect, if you want to (I decided to call this feature 'ephemerial mode').

## Features
- support for proxy (via `http_proxy`/`https_proxy` environment wariables)
- storage of all the information related to network in encrypted form (configuration, logs, etc.)
- steganography support for more latent communications (in progress)

## Installation, Deveopment & Usage
First of all, clone the repository:
```
git clone 
```

After that, build the project:
```
make release
```
And then, you are able to run the binary. Centi comes not only as a pure network program, but it also has some other features.
For example, your configuration is stored encrypted by default. This is why you are asked for password every time you are running the program.
You can create any password on startup but you need to remember it and use every time you are running the network. There is no way to disable this feature ( except messing up with a code :D ) because of security considerations.
```
$ ./centi run
```
If you are running Centi for the first time, it creates custom configuration and logs folder in `~/.centi`.
You can edit your network configuration using the following command:
```
$ ./centi editconf
```

As network logs are also stored in encrypted form (by default), you should run the following command in order to read them:
```
$ ./centi readlog
```

## Contribution
Feel free to open as issue if you have any problem with Centi. I am always open to new ideas about improving this project.

## Credits
The inspiration and main idea for this project were taken from ![this amazing person](https://github.com/number571/) and his ![project](https://github.com/number571/hidden-lake).
I also took wav parser from ![this repository](https://github.com/DylanMeeus/GoAudio).

## TODO:
- finish bluetooth module.
- think about the way to make things simplier.
- review the protocol and general security of the project (i feel like it sucks)
- improve speed of the network (as much as it possible)
- do some other optimizations and bug fixes: improve flood protection, optimize algorithms where possible
