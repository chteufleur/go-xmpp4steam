# go-xmpp4steam


go-xmpp4steam is a XMPP/Steam gateway.


## Compilation
### Dependencies

 * [go-xmpp](https://git.kingpenguin.tk/chteufleur/go-xmpp) for the XMPP part.
 * [go-steam](https://github.com/Philipp15b/go-steam) for the steam part.
 * [cfg](https://github.com/jimlawless/cfg) for the configuration file.


Go into your $GOPATH directory and execut those two line to get the 2 dependencies (cfg and go-steam).
```sh
go get github.com/Philipp15b/go-steam
go get github.com/jimlawless/cfg
```
After that, go into ``src`` directory and get the go-xmpp sources dependence.
```sh
git clone https://git.kingpenguin.tk/chteufleur/go-xmpp
```

### Download sources
Then download and compile the go-xmpp4steam gateway.
```sh
git clone https://git.kingpenguin.tk/chteufleur/go-xmpp4steam.git
cd go-xmpp4steam
go build main.go
```
A binary file will be generated.

### Configure
Configure the gateway by editing the ``xmpp4steam.cfg`` file.
The first time, let the variable ``steam_auth_code`` empty. After the first run of the gateway, Steam will send you a code that you have to give it in that variable. Then re-run the gateway and it should be OK.


## Help
To get any help, please visit the XMPP conference room at ``go-xmpp4steam@muc.kingpenguin.tk`` with your prefered client, or [with your browser](https://jappix.kingpenguin.tk/?r=go-xmpp4steam@muc.kingpenguin.tk).