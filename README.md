# go-xmpp4steam


go-xmpp4steam is a XMPP/Steam gateway.


## Compilation
### Dependencies

 * [go-xmpp](https://git.kingpenguin.tk/chteufleur/go-xmpp) for the XMPP part.
 * [go-steam](https://github.com/Philipp15b/go-steam) for the steam part.
 * [go-sqlite3](https://github.com/mattn/go-sqlite3) for the database part.
 * [cfg](https://github.com/jimlawless/cfg) for the configuration file.


### Build and run
You must first [install go environment](https://golang.org/doc/install) on your system.
Then, go into your $GOPATH directory and go get the source code (This will download the source code and the dependencies).
```sh
go get git.kingpenguin.tk/chteufleur/go-xmpp4steam.git
```

First, you need to go into directory ``$GOPATH/src/chteufleur/go-xmpp4steam.git``.
Then, you can run the project directly by using command ``go run main.go``.
Or, in order to build the project you can run the command ``go build main.go``.
It will generate a binary that you can run as any binary file.

### Configure
Configure the gateway by editing the ``xmpp4steam.conf`` file in order to give all XMPP component information. This configuration file has to be placed following the [XDG specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html) (example ``/etc/xdg/http-auth/xmpp4steam.conf``).
An example of the config file can be found in [the repos](https://git.kingpenguin.tk/chteufleur/HTTPAuthentificationOverXMPP/src/master/xmpp4steam.conf).

### Utilization
To register, you have to send an Ad-Hoc command to the gateway in order to give your Steam login information.
When it done, send a presence to the gateway. It will try to connect to Steam, but should failed.
Steam should send you a code that you have to give to the gateway using Ad-Hoc command.
After giving the code to the gateway, send again a presence to it and it should be OK.


## Help
To get any help, please visit the XMPP conference room at [go-xmpp4steam@muc.kingpenguin.tk](xmpp://go-xmpp4steam@muc.kingpenguin.tk?join) with your prefered client, or [with your browser](https://jappix.kingpenguin.tk/?r=go-xmpp4steam@muc.kingpenguin.tk).
