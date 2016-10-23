package main

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/configuration"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/database"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/gateway"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/logger"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/xmpp"

	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	Version = "v1.1-dev"
)

func init() {
	logger.Init(os.Stdout, os.Stdout, os.Stderr)
	configuration.Init()
	logger.Info.Println("Running go-xmpp4steam " + Version)
	xmpp.SoftVersion = Version

	// XMPP config
	xmpp.Addr = configuration.MapConfig["xmpp_server_address"] + ":" + configuration.MapConfig["xmpp_server_port"]
	xmpp.JidStr = configuration.MapConfig["xmpp_hostname"]
	xmpp.Secret = configuration.MapConfig["xmpp_secret"]
	xmpp.Debug = configuration.MapConfig["xmpp_debug"] == "true"
	if configuration.MapConfig["xmpp_group"] != "" {
		gateway.XmppGroupUser = configuration.MapConfig["xmpp_group"]
	}
	gateway.XmppJidComponent = xmpp.JidStr

	for _, admin := range strings.Split(configuration.MapConfig["xmpp_admins"], ";") {
		xmpp.AdminUsers[admin] = true
	}
}

func main() {
	go xmpp.Run()
	time.Sleep(1 * time.Second)
	allDbUsers := database.GetAllLines()
	for _, dbUser := range allDbUsers {
		xmpp.AddNewUser(dbUser.Jid, dbUser.SteamLogin, dbUser.SteamPwd, dbUser.Debug)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, syscall.SIGTERM)
	signal.Notify(sigchan, os.Kill)
	<-sigchan

	xmpp.Disconnect()

	logger.Info.Println("Exit main()")
	time.Sleep(1 * time.Second)
}
