package main

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/steam"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/xmpp"

	"github.com/Philipp15b/go-steam/internal/steamlang"
	"github.com/jimlawless/cfg"

	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	Version               = "go-xmpp4steam v0.1.6"
	configurationFilePath = "xmpp4steam.cfg"
)

var (
	mapConfig  = make(map[string]string)
	SetSteamId = make(map[string]struct{})
)

func init() {
	err := cfg.Load(configurationFilePath, mapConfig)
	if err != nil {
		log.Fatal("Failed to load configuration file.", err)
	}

	// XMPP config
	xmpp.Addr = mapConfig["xmpp_server_address"] + ":" + mapConfig["xmpp_server_port"]
	xmpp.JidStr = mapConfig["xmpp_hostname"]
	xmpp.Secret = mapConfig["xmpp_secret"]
	xmpp.PreferedJID = mapConfig["xmpp_authorized_jid"]
	xmpp.Debug = mapConfig["xmpp_debug"] == "true"

	// Steam config
	steam.Username = mapConfig["steam_login"]
	steam.Password = mapConfig["steam_password"]
	steam.AuthCode = ""
}

func main() {
	go gatewayXmppSteamAction()
	go gatewaySteamXmppAction()

	go gatewayXmppSteamPresence()
	go gatewayXmppSteamMessage()
	go gatewayXmppSteamAuthCode()

	go gatewaySteamXmppMessage()
	go gatewaySteamXmppPresence()

	go steam.Run()
	go xmpp.Run()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, syscall.SIGTERM)
	signal.Notify(sigchan, os.Kill)
	<-sigchan

	steam.Disconnect()
	xmpp.Disconnect()

	time.Sleep(1 * time.Second)
	log.Println("Exit main()")
}

// XMPP -> Steam gateways
func gatewayXmppSteamAction() {
	for {
		action := <-xmpp.ChanAction

		switch action {
		case xmpp.ActionConnexion:
			if !steam.IsConnected() {
				steam.Connect()
			}

		case xmpp.ActionDeconnexion:
			if steam.IsConnected() {
				steam.Disconnect()
			}

		case xmpp.ActionMainMethodEnded:
			go xmpp.Run()
		}
	}
}

func gatewayXmppSteamPresence() {
	for {
		status := <-xmpp.ChanPresence
		tpye := <-xmpp.ChanPresence

		var steamStatus steamlang.EPersonaState

		switch status {
		case xmpp.Status_online:
			steamStatus = steam.State_Online

		case xmpp.Status_away:
			steamStatus = steam.State_Away

		case xmpp.Status_chat:

		case xmpp.Status_extended_away:
			steamStatus = steam.State_Snooze

		case xmpp.Status_do_not_disturb:
			steamStatus = steam.State_Busy
		}

		steam.SendPresence(steamStatus)
		xmpp.SendPresence(status, tpye, Version)
	}
}

func gatewayXmppSteamMessage() {
	for {
		steamId := <-xmpp.ChanMessage
		message := <-xmpp.ChanMessage

		steam.SendMessage(steamId, message)
	}
}

func gatewayXmppSteamAuthCode() {
	for {
		authCode := <-xmpp.ChanAuthCode
		steam.AuthCode = authCode
		steam.Disconnect()
		time.Sleep(2 * time.Second)
		go steam.Run()
	}
}

// /XMPP -> Steam gateways

// Steam -> XMPP gateways
func gatewaySteamXmppAction() {
	for {
		action := <-steam.ChanAction
		switch action {
		case steam.ActionConnected:
			xmpp.SendPresence(xmpp.CurrentStatus, xmpp.Type_available, Version)

		case steam.ActionDisconnected:
			xmpp.Disconnect()
			disconnectAllSteamUser()

		case steam.ActionFatalError:
			disconnectAllSteamUser()
			time.Sleep(2 * time.Second)
			go steam.Run()

		case steam.ActionMainMethodEnded:
			go steam.Run()
		}
	}
}

func gatewaySteamXmppMessage() {
	for {
		steamId := <-steam.ChanMessage
		message := <-steam.ChanMessage
		xmpp.SendMessage(steamId+"@"+xmpp.JidStr, message)
	}
}

func gatewaySteamXmppPresence() {
	for {
		steamId := <-steam.ChanPresence
		name := <-steam.ChanPresence
		stat := <-steam.ChanPresenceSteam
		gameName := <-steam.ChanPresence

		var status string
		var tpye string
		switch stat {
		case steam.State_Offline:
			status = xmpp.Status_offline
			tpye = xmpp.Type_unavailable

		case steam.State_Online:
			status = xmpp.Status_online
			tpye = xmpp.Type_available

		case steam.State_Busy:
			status = xmpp.Status_do_not_disturb
			tpye = xmpp.Type_available

		case steam.State_Away:
			status = xmpp.Status_away
			tpye = xmpp.Type_available

		case steam.State_Snooze:
			status = xmpp.Status_extended_away
			tpye = xmpp.Type_available
		}

		if _, ok := SetSteamId[steamId]; !ok {
			xmpp.SendPresenceFrom(status, xmpp.Type_subscribe, steamId+"@"+xmpp.JidStr, gameName, name)
			SetSteamId[steamId] = struct{}{}
		}
		xmpp.SendPresenceFrom(status, tpye, steamId+"@"+xmpp.JidStr, gameName, name)
	}
}

func disconnectAllSteamUser() {
	for sid, _ := range SetSteamId {
		xmpp.SendPresenceFrom(xmpp.Status_offline, xmpp.Type_unavailable, sid+"@"+xmpp.JidStr, "", "")
		delete(SetSteamId, sid)
	}
}

// /Steam -> XMPP gateways
