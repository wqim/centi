# Centi

## About project
Centi is a doomsday hidden network which can work with any way of communications: from mesh-networks to centralized platforms.
Centi is a hybrid network. It is able to have any topology. Need to build an anonymous network over a centralized service?
No problem, just write a microservice for it's API. Over a decentralized network? You are welcome. As a distributed network?
No problems, dude.

>[!WARNING]
>This is an experimental project. Do not rely on Centi for important actions.

## Current state & development
The project is in active development. Be aware of breaking changes which can be introduced without warnings.

## How does it work?
As concepts introduced in such networks as Tor and I2P aren't suitable for (almost?) censured internet,
the core idea of anonymisation mechanism was taken from HiddenLake.
So, receiver is anonymized by other network participants by sending every packet to every known participant of the network.
Sender is anonymized by generation of random packets and sending them with randomized time delays. The only reason delays
are randomized is not to trigger some kind of WAF or EDR on platforms of big corporations. This behaviour may be changed in
configuration.

In order to make network work on any platform with any way of communication, Centi uses microservices which provide
API for sending and receiving data, initializing and deleting channels of communications (e.g. initializing repository
in GitHub which you will use only for communicating inside the network) and some other things (compile `docs/modules.tex`
to PDF for more information about this topic).

## Available microservices
- TLS
- SSH (still not always working but is here)
- Reticulum
- Bluetooth (in process)
- GitHub, Gitea and HuggingFace platforms
- Email (also in process)

## Features
- support for proxy (via `http_proxy`/`https_proxy` environment variables)
- support for usage of GitHub, Gitea, HuggingFace, TLS connections and Reticulum out of the box.
- storage of all the information related to network in encrypted form (configuration, logs, etc.)
- temporary files are shredded after usage.
- steganography support for more silent communications

## Requirements
- golang 1.24 or higher (for building application)
- python3 (for some microservices)
- docker (optionally)

## Installation & Usage
First of all, clone the repository in `$GOPATH/src` directory:
```
git clone https://github.com/wqim/centi
```

After that, build the project:
```
make release
```
And then, you are able to run the binary. Centi comes not only as a pure network program, it also has some other features.
For example, your configuration is encrypted by default. This is why you are asked for password every time you are running the program.
You can create any password on startup but you need to remember it. There is no way to disable this feature ( except messing up with a code :D ) because of security considerations.
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

## Configuration
Some parts of your configuration are generating at first run. Centi generates random password for database, API server configuration, queue and packet sizes
and some other default stuff.
Other parts of configuration aren't generated automatically. You must write it in order to connect to network in the desired way.
By default, Centi generates configuration similar to one below:

```yaml
network_config:
    min_delay: 10000
    max_delay: 20000
    keys_collection_delay: 5000
    queue_size: 100
    packet_size: 4096
    accept_unknown: true
    send_known_peers: true
    network_key: ""
    peers: []
local_server_config:
    address: 127.0.0.1:8080
    not_found_page: www/404.html
    pages:
        GET /{$}: www/index.html
        GET /script.js: www/script.js
        GET /styles.css: www/styles.css
steganography_config:
    decoy_files_folder: ""
logger_config:
    filename: $HOME/.centi/log.log
    password: ""
    isencrypted: false
    iscolored: true
    savetime: true
    mode: 1
platforms_data:
    - platform: universal
      args:
        addr: http://127.0.0.1:9000
        autodiscovery: "true"
        config_path: $HOME/.reticulum
        max_attempts: "10000"
        name: reticulum
        run_as_server: "true"
      channels: []
    - platform: universal
      args:
        addr: http://127.0.0.1:3333
        autodiscovery: "true"
        name: bluetooth
        run_as_server: "true"
      channels: []
db_file: $HOME/.centi/db.db
db_password: <your-database-password-here>
db_rows_limit: 10000
keys:
    public_key: <your-public-key-here>
    private_key: <your-private-key-here>
```

As you can understand from this code, Centi uses Reticulum and Bluetooth microservices for automatic peer discovery and connection.
In order to connect to someone, you need to know their public key. This behvaiour may change in the future if reliable enough
mechanism of public key transmission will be found.
You can send data only to those to whom you have connected. "Connect" word here means what you have sent a ciphertext (Kyber768 part of KEM) and 
ephemerial public key (Ed25519 part) to someone, who's public key is known to you and you and your peer have the same shared secret now.
 
Simple TLS-channeled network configuration looks like this:
```yaml
# ...
platforms_data:
    # ... other platforms data here
    - platform: tls
      args:
        run_server: "true"
        packet_size: "4096" # your packet size in form of the string here
        protocol: "tcp" # just to make things work correct
        cert_path: "<path-to-your-server-node-certificate>"
        key_path: "<path-to-your-server-node-key>"
        net_addr: "127.0.0.1:9000" # address to bind server on
        max_connections: 5 # maximum amount of connections accepted by server
      channels:
        - name: "127.0.0.1:9001" # an address of server node to connect to
          args: # this field isn't used by tls module but some other modules can utilize this field to store
            # information about supported files formats ("steganographied" part), API tokens and other credentials.
            any_argument_name_here: "any_value_here"
# ...
```

Every module uses the same construction for it's configuration. It's form supplied below.
```yaml
# ...
platform_data:
    - platform: <module_name>
      args:
        arg1: value1
        # ...
      channels:
        - name: channel1
          args:
            arg1: value1
            # ...
        # ...
# ...
```

If you have developed or found a suitable microservice, you can use `universal` module to connect this microservice. The example is
the default configuration with Reticulum and Bluetooth microservices.

## Contribution
Feel free to open as issue if you have any problem with Centi. I am always open to new ideas about improving this project.

## Credits
The inspiration and main idea for this project were taken from ![this amazing person](https://github.com/number571/) and his ![project](https://github.com/number571/hidden-lake).
I also took wav parser from ![this repository](https://github.com/DylanMeeus/GoAudio).

## TODO:
- finish bluetooth microservice
- do some optimizations and bug fixes: improve flood protection, optimize algorithms where possible, remove unused parts of code
