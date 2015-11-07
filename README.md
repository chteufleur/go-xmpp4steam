# go-xmpp4steam


go-xmpp4steam is a XMPP/Steam gateway.


## Compilation
### Dependencies

 * [go-xmpp](https://git.kingpenguin.tk/chteufleur/go-xmpp) for the XMPP part.
 * [go-steam](https://github.com/Philipp15b/go-steam) for the steam part.
 * [cfg](https://github.com/jimlawless/cfg) for the configuration file.


Download the CA at [https://kingpenguin.tk/ressources/cacert.pem](https://kingpenguin.tk/ressources/cacert.pem), then edit your .gitconfig and add the following lines
```
[https "https://git.kingpenguin.tk"]
  sslCAPath = /path/to/CA
```

Then go into your $GOPATH directory and go get the code.
```sh
go get git.kingpenguin.tk/chteufleur/go-xmpp4steam.git
```

### Configure
Configure the gateway by editing the ``xmpp4steam.cfg`` file.
The first time, let the variable ``steam_auth_code`` empty. After the first run of the gateway, Steam will send you a code that you have to give it in that variable. Then re-run the gateway and it should be OK.


## Help
To get any help, please visit the XMPP conference room at ``go-xmpp4steam@muc.kingpenguin.tk`` with your prefered client, or [with your browser](https://jappix.kingpenguin.tk/?r=go-xmpp4steam@muc.kingpenguin.tk).
