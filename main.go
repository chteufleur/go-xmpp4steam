package main

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/database"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/gateway"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/xmpp"

	"github.com/jimlawless/cfg"

	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	Version               = "v1.1-dev"
	configurationFilePath = "xmpp4steam.cfg"
)

var (
	mapConfig = make(map[string]string)
)

func init() {
	xmpp.SoftVersion = Version
	err := cfg.Load(configurationFilePath, mapConfig)
	if err != nil {
		log.Fatal("Failed to load configuration file.", err)
	}

	// XMPP config
	xmpp.Addr = mapConfig["xmpp_server_address"] + ":" + mapConfig["xmpp_server_port"]
	xmpp.JidStr = mapConfig["xmpp_hostname"]
	xmpp.Secret = mapConfig["xmpp_secret"]
	xmpp.Debug = mapConfig["xmpp_debug"] == "true"
	gateway.XmppJidComponent = xmpp.JidStr

	os.MkdirAll(gateway.SentryDirectory, 0700)
}

func main() {
	allDbUsers := database.GetAllLines()
	for _, dbUser := range allDbUsers {
		xmpp.AddNewUser(dbUser.Jid, dbUser.SteamLogin, dbUser.SteamPwd, dbUser.Debug)
	}
	go xmpp.Run()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, syscall.SIGTERM)
	signal.Notify(sigchan, os.Kill)
	<-sigchan

	xmpp.Disconnect()

	log.Println("Exit main()")
	time.Sleep(1 * time.Second)
}
