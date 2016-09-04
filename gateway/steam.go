package gateway

import (
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/Philipp15b/go-steam/steamid"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"
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
	if g.Deleting {
		log.Printf("%sDeleting gateway", LogSteamInfo)
		return
	}

	log.Printf("%sRunning", LogSteamInfo)
	g.setLoginInfos()
	g.SteamClient = steam.NewClient()
	g.SteamConnecting = false
	g.SteamClient.ConnectionTimeout = 10 * time.Second

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
			if _, ok := g.SteamClient.Social.Friends.GetCopy()[e.FriendId]; !ok {
				// Is not in friend list
				// Exepte for myself
				if g.SteamClient.SteamId() != e.FriendId {
					continue
				}
			}
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
				g.SendXmppPresence(status, Type_subscribe, "", steamId+"@"+XmppJidComponent, gameName, name)
				g.FriendSteamId[steamId] = &StatusSteamFriend{XMPP_Status: status, XMPP_Type: tpye}
			} else {
				g.FriendSteamId[steamId].XMPP_Status = status
				g.FriendSteamId[steamId].XMPP_Type = tpye
				g.FriendSteamId[steamId].SteamGameName = gameName
				g.FriendSteamId[steamId].SteamName = name
			}
			g.SendXmppPresence(status, tpye, "", steamId+"@"+XmppJidComponent, gameName, name)

		case *steam.ChatMsgEvent:
			// Message received
			if e.EntryType == steamlang.EChatEntryType_Typing {
				g.SendXmppMessageComposing(e.ChatterId.ToString() + "@" + XmppJidComponent)
			} else {
				g.SendXmppMessage(e.ChatterId.ToString()+"@"+XmppJidComponent, "", e.Message)
			}

		case *steam.ChatInviteEvent:
			// Invitation to play
			if fromFriend, ok := g.SteamClient.Social.Friends.GetCopy()[e.FriendChatId]; ok {
				messageToSend := fmt.Sprintf("Currently playing to « %s », would you like to join ?", fromFriend.GameName)
				g.SendXmppMessage(e.FriendChatId.ToString()+"@"+XmppJidComponent, "", messageToSend)
			}

		default:
			log.Printf("%s", LogSteamDebug, e)
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
	}
	log.Printf("%sAuthentification of (%s, %s)", LogSteamDebug, g.XMPP_JID_Client, g.SteamLoginInfo.Username)
}

func (g *GatewayInfo) IsSteamConnected() bool {
	ret := g != nil
	if ret {
		ret = g.SteamClient != nil
		if ret {
			ret = g.SteamClient.Connected()
		}
	}
	return ret
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
		g.SendXmppPresence(Status_offline, Type_unavailable, "", sid+"@"+XmppJidComponent, "", "")
		delete(g.FriendSteamId, sid)
	}
}

func (g *GatewayInfo) SendSteamMessage(steamId, message string) {
	g.sendSteamMessage(steamId, message, steamlang.EChatEntryType_ChatMsg)
}

func (g *GatewayInfo) SendSteamMessageComposing(steamId string) {
	g.sendSteamMessage(steamId, "", steamlang.EChatEntryType_Typing)
}

func (g *GatewayInfo) sendSteamMessage(steamId, message string, chatEntryType steamlang.EChatEntryType) {
	if !g.IsSteamConnected() {
		log.Printf("%sTry to send message, but disconnected", LogSteamDebug)
		return
	}

	steamIdUint64, err := strconv.ParseUint(steamId, 10, 64)
	if err == nil {
		g.SteamClient.Social.SendMessage(steamid.SteamId(steamIdUint64), chatEntryType, message)
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
