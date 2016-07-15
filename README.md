# go-xmpp4steam


go-xmpp4steam is a XMPP/Steam gateway.


## Compilation
### Dependencies

 * [go-xmpp](https://git.kingpenguin.tk/chteufleur/go-xmpp) for the XMPP part.
 * [go-steam](https://github.com/Philipp15b/go-steam) for the steam part.
 * [go-sqlite3](https://github.com/mattn/go-sqlite3) for the database part.
 * [cfg](https://github.com/jimlawless/cfg) for the configuration file.


Download the CA at [https://kingpenguin.tk/ressources/cacert.pem](https://kingpenguin.tk/ressources/cacert.pem), then install it on your operating system.
Once installed, go into your $GOPATH directory and go get the source code.
```sh
go get git.kingpenguin.tk/chteufleur/go-xmpp4steam.git
```

### Configure
Configure the gateway by editing the ``xmpp4steam.cfg`` file in order to give all XMPP component information.

### Utilization
To register, you have to send an Ad-Hoc command to the gateway in order to give your Steam login information.
When it done, send a presence to the gateway. It will try to connect to Steam, but should failed.
Steam should send you a code that you have to give to the gateway using Ad-Hoc command.
After giving the code to the gateway, send again a presence to it and it should be OK.


## Help
To get any help, please visit the XMPP conference room at [go-xmpp4steam@muc.kingpenguin.tk](xmpp://go-xmpp4steam@muc.kingpenguin.tk?join) with your prefered client, or [with your browser](https://jappix.kingpenguin.tk/?r=go-xmpp4steam@muc.kingpenguin.tk).
