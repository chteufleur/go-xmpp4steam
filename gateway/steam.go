package gateway

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/logger"
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/protocol/steamlang"
	"github.com/Philipp15b/go-steam/steamid"

	"fmt"
	"io/ioutil"
	"strconv"
	"time"
)

const (
	State_Offline        = steamlang.EPersonaState_Offline
	State_Online         = steamlang.EPersonaState_Online
	State_Busy           = steamlang.EPersonaState_Busy
	State_Away           = steamlang.EPersonaState_Away
	State_Snooze         = steamlang.EPersonaState_Snooze
	State_LookingToTrade = steamlang.EPersonaState_LookingToTrade
	State_LookingToPlay  = steamlang.EPersonaState_LookingToPlay
	State_Max            = steamlang.EPersonaState_Max
)

var (
	ServerAddrs = "servers.addr"
)

func (g *GatewayInfo) SteamRun() {
	if g.Deleting {
		logger.Info.Printf("[%s] Deleting gateway", g.XMPP_JID_Client)
		return
	}

	logger.Info.Printf("[%s] Running", g.XMPP_JID_Client)
	steam.InitializeSteamDirectory()
	g.setLoginInfos()
	if g.SteamClient == nil {
		g.SteamClient = steam.NewClient()
	}
	g.SteamConnecting = false
	g.SteamClient.ConnectionTimeout = 10 * time.Second

	g.mainSteam()

	logger.Info.Printf("[%s] Reach main method's end", g.XMPP_JID_Client)
}

func (g *GatewayInfo) mainSteam() {
	for event := range g.SteamClient.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			// Connected on server
			g.SteamConnecting = false
			logger.Debug.Printf("[%s] Connected on Steam serveur", g.XMPP_JID_Client)
			g.SteamClient.Auth.LogOn(g.SteamLoginInfo)

		case *steam.MachineAuthUpdateEvent:
			// Received sentry file
			ioutil.WriteFile(g.SentryFile, e.Hash, 0666)

		case *steam.LoggedOnEvent:
			// Logged on
			g.SendSteamPresence(steamlang.EPersonaState_Online)
			g.SendXmppMessage(XmppJidComponent, "", "Connected on Steam network")

		case *steam.LoggedOffEvent:
			logger.Error.Printf("[%s] LoggedOffEvent: %v", g.XMPP_JID_Client, e)
			g.SendXmppMessage(XmppJidComponent, "", fmt.Sprintf("Disconnected of Steam network (%v)", e))
			g.SteamConnecting = false

		case steam.FatalErrorEvent:
			logger.Error.Printf("[%s] FatalError: %v", g.XMPP_JID_Client, e)
			g.SendXmppMessage(XmppJidComponent, "", fmt.Sprintf("Steam Fatal Error : %v", e))
			g.DisconnectAllSteamFriend()
			g.SteamConnecting = false

		case *steam.DisconnectedEvent:
			logger.Info.Printf("[%s] Disconnected event", g.XMPP_JID_Client)
			g.SendXmppMessage(XmppJidComponent, "", fmt.Sprintf("Steam Error : %v", e))
			g.DisconnectAllSteamFriend()
			g.SteamConnecting = false

		case error:
			logger.Error.Printf("[%s] error: %v", g.XMPP_JID_Client, e)
			g.SendXmppMessage(XmppJidComponent, "", "Steam Error : "+e.Error())

		case *steam.LogOnFailedEvent:
			logger.Error.Printf("[%s] Login failed: %v", g.XMPP_JID_Client, e)
			g.SendXmppMessage(XmppJidComponent, "", fmt.Sprintf("Login failed : %v", e.Result))
			g.SteamConnecting = false

		case *steam.ClientCMListEvent:
			// Doing nothing with server list

		case *steam.PersonaStateEvent:
			logger.Debug.Printf("[%s] Received PersonaStateEvent: %v", g.XMPP_JID_Client, e)
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
			logger.Debug.Printf("[%s] Received ChatMsgEvent: %v", g.XMPP_JID_Client, e)
			// Message received
			from := e.ChatterId.ToString() + "@" + XmppJidComponent
			if e.EntryType == steamlang.EChatEntryType_Typing {
				g.SendXmppMessageComposing(from)
			} else if e.EntryType == steamlang.EChatEntryType_LeftConversation {
				g.SendXmppMessageLeaveConversation(from)
			} else {
				g.SendXmppMessage(from, "", e.Message)
			}

		case *steam.ChatInviteEvent:
			logger.Debug.Printf("[%s] Received ChatInviteEvent: %v", g.XMPP_JID_Client, e)
			// Invitation to play
			if fromFriend, ok := g.SteamClient.Social.Friends.GetCopy()[e.FriendChatId]; ok {
				messageToSend := fmt.Sprintf("Currently playing to « %s », would you like to join ?", fromFriend.GameName)
				g.SendXmppMessage(e.FriendChatId.ToString()+"@"+XmppJidComponent, "", messageToSend)
			}

		default:
			logger.Debug.Printf("[%s] Steam unmatch event (Type: %T): %v", g.XMPP_JID_Client, e, e)
			g.SendXmppMessage(XmppJidComponent, "", fmt.Sprintf("Steam unmatch event (Type: %T): %v", e, e))
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
	logger.Debug.Printf("Authentification of (%s, %s)", g.XMPP_JID_Client, g.SteamLoginInfo.Username)
}

func (g *GatewayInfo) IsSteamConnected() bool {
	ret := g != nil
	if ret {
		ret = g.SteamClient != nil
		if ret {
			ret = g.SteamClient.Connected()
		}
	}
	logger.Debug.Printf("[%s] Is Steam connected (Connected: %v)", g.XMPP_JID_Client, ret)
	return ret
}

func (g *GatewayInfo) SteamConnect() {
	if g.IsSteamConnected() {
		logger.Debug.Printf("[%s] Try to connect, but already connected", g.XMPP_JID_Client)
		return
	}
	if g.SteamConnecting {
		logger.Debug.Printf("[%s] Try to connect, but currently connecting…", g.XMPP_JID_Client)
		return
	}

	g.SteamConnecting = true
	go func() {
		logger.Info.Printf("[%s] Connecting...", g.XMPP_JID_Client)
		g.SendXmppMessage(XmppJidComponent, "", "Connecting...")
		addr := g.SteamClient.Connect()
		logger.Info.Printf("[%s] Connected on %v", g.XMPP_JID_Client, addr)
		g.SendXmppMessage(XmppJidComponent, "", fmt.Sprintf("Connected on %v", addr))
	}()
}

func (g *GatewayInfo) SteamDisconnect() {
	if !g.IsSteamConnected() {
		logger.Debug.Printf("[%s] Try to disconnect, but already disconnected", g.XMPP_JID_Client)
		return
	}
	logger.Info.Printf("[%s] Steam disconnect", g.XMPP_JID_Client)

	g.XMPP_Disconnect()
	g.DisconnectAllSteamFriend()
	go g.SteamClient.Disconnect()
}

func (g *GatewayInfo) DisconnectAllSteamFriend() {
	logger.Debug.Printf("[%s] Disconnect all Steam friend", g.XMPP_JID_Client)
	for sid := range g.FriendSteamId {
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

func (g *GatewayInfo) SendSteamMessageLeaveConversation(steamId string) {
	g.sendSteamMessage(steamId, "", steamlang.EChatEntryType_LeftConversation)
}

func (g *GatewayInfo) sendSteamMessage(steamId, message string, chatEntryType steamlang.EChatEntryType) {
	if !g.IsSteamConnected() {
		logger.Debug.Printf("[%s] Try to send message, but disconnected", g.XMPP_JID_Client)
		return
	}

	steamIdUint64, err := strconv.ParseUint(steamId, 10, 64)
	if err == nil {
		logger.Debug.Printf("[%s] Send message to %v", g.XMPP_JID_Client, steamIdUint64)
		g.SteamClient.Social.SendMessage(steamid.SteamId(steamIdUint64), chatEntryType, message)
	} else {
		logger.Error.Printf("[%s] Failed to get SteamId from %s", g.XMPP_JID_Client, steamId)
	}
}

func (g *GatewayInfo) SendSteamPresence(status steamlang.EPersonaState) {
	if !g.IsSteamConnected() {
		logger.Debug.Printf("[%s] Try to send presence, but disconnected", g.XMPP_JID_Client)
		return
	}
	logger.Debug.Printf("[%s] Send presence (Status: %v)", g.XMPP_JID_Client, status)
	g.SteamClient.Social.SetPersonaState(status)
}
