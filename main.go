package main

import (
	"go-xmpp4steam/steam"
	"go-xmpp4steam/xmpp"

	"github.com/Philipp15b/go-steam/internal/steamlang"
	"github.com/jimlawless/cfg"

	"bufio"
	"log"
	"os"
	"strings"
	"time"
)

const (
	configurationFilePath = "xmpp4steam.cfg"
)

var (
	mapConfig  = make(map[string]string)
	SetSteamId = make(map[string]struct{})
)

func init() {
	xmpp.Version = "0.1.1"

	err := cfg.Load(configurationFilePath, mapConfig)
	if err != nil {
		log.Fatal("Failed to load configuration file.", err)
	}

	// XMPP config
	xmpp.Addr = mapConfig["xmpp_server_address"] + ":" + mapConfig["xmpp_server_port"]
	xmpp.JidStr = mapConfig["xmpp_hostname"]
	xmpp.Secret = mapConfig["xmpp_secret"]
	xmpp.PreferedJID = mapConfig["xmpp_authorized_jid"]

	// Steam config
	steam.Username = mapConfig["steam_login"]
	steam.Password = mapConfig["steam_password"]
	steam.AuthCode = mapConfig["steam_auth_code"]
}

func main() {
	go gatewayXmppSteamAction()
	go gatewaySteamXmppAction()

	go gatewayXmppSteamPresence()
	go gatewayXmppSteamMessage()

	go gatewaySteamXmppMessage()
	go gatewaySteamXmppPresence()

	go steam.Run()
	xmpp.Run()

	// inputStop()

	steam.Disconnect()
	xmpp.Disconnect()
	time.Sleep(1 * time.Second)
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
		}
	}
}

func gatewayXmppSteamPresence() {
	for {
		status := <-xmpp.ChanPresence
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
	}
}

func gatewayXmppSteamMessage() {
	for {
		steamId := <-xmpp.ChanMessage
		message := <-xmpp.ChanMessage

		steam.SendMessage(steamId, message)
	}
}

// /XMPP -> Steam gateways

// Steam -> XMPP gateways
func gatewaySteamXmppAction() {
	for {
		action := <-steam.ChanAction
		switch action {
		case steam.ActionConnected:
			xmpp.SendPresence(xmpp.CurrentStatus, xmpp.Type_available)

		case steam.ActionDisconnected:
			xmpp.Disconnect()
			for sid, _ := range SetSteamId {
				xmpp.SendPresenceFrom(xmpp.Status_offline, xmpp.Type_unavailable, sid+"@"+xmpp.JidStr, "")
				delete(SetSteamId, sid)
			}

		case steam.ActionFatalError:
			time.Sleep(2 * time.Second)
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
		// name := steam.ChanPresence
		steamId := <-steam.ChanPresence
		stat := <-steam.ChanPresenceSteam
		gameName := <-steam.ChanPresence

		SetSteamId[steamId] = struct{}{}

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

		xmpp.SendPresenceFrom(status, tpye, steamId+"@"+xmpp.JidStr, gameName)
	}
}

// /Steam -> XMPP gateways

func inputStop() {
	for {
		in := bufio.NewReader(os.Stdin)
		line, err := in.ReadString('\n')
		if err != nil {
			continue
		}
		line = strings.TrimRight(line, "\n")

		if line == "stop" {
			return
		}
	}
}
