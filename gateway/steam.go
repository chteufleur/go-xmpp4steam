package gateway

import (
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/internal/steamlang"
	"github.com/Philipp15b/go-steam/steamid"

	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	// "time"
)

const (
	serverAddrs = "servers.addr"

	State_Offline        = steamlang.EPersonaState_Offline
	State_Online         = steamlang.EPersonaState_Online
	State_Busy           = steamlang.EPersonaState_Busy
	State_Away           = steamlang.EPersonaState_Away
	State_Snooze         = steamlang.EPersonaState_Snooze
	State_LookingToTrade = steamlang.EPersonaState_LookingToTrade
	State_LookingToPlay  = steamlang.EPersonaState_LookingToPlay
	State_Max            = steamlang.EPersonaState_Max

	LogSteamInfo  = "\t[STEAM INFO]\t"
	LogSteamError = "\t[STEAM ERROR]\t"
	LogSteamDebug = "\t[STEAM DEBUG]\t"
)

func (g *GatewayInfo) SteamRun() {
	log.Printf("%sRunning", LogSteamInfo)
	g.setLoginInfos()
	g.SteamClient = steam.NewClient()
	g.SteamConnecting = false
	// g.SteamClient.ConnectionTimeout = 10 * time.Second

	g.mainSteam()

	log.Printf("%sReach main method's end", LogSteamInfo)
	go g.SteamRun()
}

func (g *GatewayInfo) mainSteam() {
	for event := range g.SteamClient.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			// Connected on server
			g.SteamConnecting = false
			log.Printf("%sConnected on Steam serveur", LogSteamDebug)
			g.SteamClient.Auth.LogOn(g.SteamLoginInfo)

		case *steam.MachineAuthUpdateEvent:
			// Received sentry file
			ioutil.WriteFile(g.SentryFile, e.Hash, 0666)

		case *steam.LoggedOnEvent:
			// Logged on
			g.SendSteamPresence(steamlang.EPersonaState_Online)
			g.SendXmppMessage(XmppJidComponent, "", "Connected on Steam network")

		case steam.FatalErrorEvent:
			log.Printf("%sFatalError: ", LogSteamError, e)
			g.SendXmppMessage(XmppJidComponent, "", "Steam Fatal Error : "+e.Error())
			g.DisconnectAllSteamFriend()
			return

		case error:
			log.Printf("%s", LogSteamError, e)
			g.SendXmppMessage(XmppJidComponent, "", "Steam Error : "+e.Error())

		case *steam.ClientCMListEvent:
			// Save servers addresses
			b, err := json.Marshal(*e)
			if err != nil {
				log.Printf("%sFailed to json.Marshal() servers list", LogSteamError)
			} else {
				ioutil.WriteFile(serverAddrs, b, 0666)
			}

		case *steam.PersonaStateEvent:
			// Presenc received
			steamId := e.FriendId.ToString()
			name := e.Name
			gameName := e.GameName

			var status string
			var tpye string
			switch e.State {
			case State_Offline:
				status = Status_offline
				tpye = Type_unavailable
			case State_Online:
				status = Status_online
				tpye = Type_available
			case State_Busy:
				status = Status_do_not_disturb
				tpye = Type_available
			case State_Away:
				status = Status_away
				tpye = Type_available
			case State_Snooze:
				status = Status_extended_away
				tpye = Type_available
			}
			if _, ok := g.FriendSteamId[steamId]; !ok {
				// Send subscribsion
				g.SendXmppPresence(status, Type_subscribe, steamId+"@"+XmppJidComponent, gameName, name)
				g.FriendSteamId[steamId] = struct{}{}
			}
			g.SendXmppPresence(status, tpye, steamId+"@"+XmppJidComponent, gameName, name)

		case *steam.ChatMsgEvent:
			// Message received
			g.SendXmppMessage(e.ChatterId.ToString()+"@"+XmppJidComponent, "", e.Message)

		default:
			log.Printf("%s", LogSteamDebug, e)
			// TODO send message
		}
	}
}

func (g *GatewayInfo) setLoginInfos() {
	var sentryHash steam.SentryHash
	sentryHash, err := ioutil.ReadFile(g.SentryFile)

	g.SteamLoginInfo = new(steam.LogOnDetails)
	g.SteamLoginInfo.Username = g.SteamLogin
	g.SteamLoginInfo.Password = g.SteamPassword

	if err == nil {
		g.SteamLoginInfo.SentryFileHash = sentryHash
		log.Printf("%sAuthentification by SentryFileHash", LogSteamDebug)
	} else if g.SteamAuthCode != "" {
		g.SteamLoginInfo.AuthCode = g.SteamAuthCode
		log.Printf("%sAuthentification by AuthCode (%s, %s, %s)", LogSteamDebug, g.SteamLoginInfo.Username, g.SteamLoginInfo.Password, g.SteamAuthCode)
	} else {
		log.Printf("%sFirst authentification (%s, %s)", LogSteamDebug, g.SteamLoginInfo.Username, g.SteamLoginInfo.Password)
	}
}

func (g *GatewayInfo) IsSteamConnected() bool {
	return g.SteamClient.Connected()
}

func (g *GatewayInfo) SteamConnect() {
	if g.IsSteamConnected() {
		log.Printf("%sTry to connect, but already connected", LogSteamDebug)
		return
	}
	if g.SteamConnecting {
		log.Printf("%sTry to connect, but currently connecting…", LogSteamDebug)
		return
	}

	g.SteamConnecting = true
	b, err := ioutil.ReadFile(serverAddrs)
	if err == nil {
		var toList steam.ClientCMListEvent
		err := json.Unmarshal(b, &toList)
		if err != nil {
			log.Printf("%sFailed to json.Unmarshal() servers list", LogSteamError)
		} else {
			log.Printf("%sConnecting...", LogSteamInfo, toList.Addresses[0])
			g.SteamClient.ConnectTo(toList.Addresses[0])
		}
	} else {
		log.Printf("%sFailed to read servers list file", LogSteamError)
		log.Printf("%sConnecting...", LogSteamInfo)
		g.SteamClient.Connect()
	}
}

func (g *GatewayInfo) SteamDisconnect() {
	if !g.IsSteamConnected() {
		log.Printf("%sTry to disconnect, but already disconnected", LogSteamDebug)
		return
	}
	log.Printf("%sSteam disconnect", LogSteamInfo)

	g.XMPP_Disconnect()
	g.DisconnectAllSteamFriend()
	go g.SteamClient.Disconnect()
}

func (g *GatewayInfo) DisconnectAllSteamFriend() {
	for sid, _ := range g.FriendSteamId {
		g.SendXmppPresence(Status_offline, Type_unavailable, sid+"@"+XmppJidComponent, "", "")
		delete(g.FriendSteamId, sid)
	}
}

func (g *GatewayInfo) SendSteamMessage(steamId, message string) {
	if !g.IsSteamConnected() {
		log.Printf("%sTry to send message, but disconnected", LogSteamDebug)
		return
	}

	steamIdUint64, err := strconv.ParseUint(steamId, 10, 64)
	if err == nil {
		g.SteamClient.Social.SendMessage(steamid.SteamId(steamIdUint64), steamlang.EChatEntryType_ChatMsg, message)
	} else {
		log.Printf("%sFailed to get SteamId from %s", LogSteamError, steamId)
	}
}

func (g *GatewayInfo) SendSteamPresence(status steamlang.EPersonaState) {
	if !g.IsSteamConnected() {
		log.Printf("%sTry to send presence, but disconnected", LogSteamDebug)
		return
	}
	g.SteamClient.Social.SetPersonaState(status)
}
