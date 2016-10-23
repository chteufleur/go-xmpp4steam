package gateway

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp.git/src/xmpp"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/logger"
	"github.com/Philipp15b/go-steam/protocol/steamlang"

	"strconv"
	"strings"
	"time"
)

const (
	Status_online         = ""
	Status_offline        = ""
	Status_away           = "away"
	Status_chat           = "chat"
	Status_do_not_disturb = "dnd"
	Status_extended_away  = "xa"

	Type_available    = ""
	Type_unavailable  = "unavailable"
	Type_subscribe    = "subscribe"
	Type_subscribed   = "subscribed"
	Type_unsubscribe  = "unsubscribe"
	Type_unsubscribed = "unsubscribed"
	Type_probe        = "probe"
	Type_error        = "error"

	ActionConnexion       = "action_xmpp_connexion"
	ActionDeconnexion     = "action_xmpp_deconnexion"
	ActionMainMethodEnded = "action_xmpp_main_method_ended"
)

var (
	XmppJidComponent = ""

	iqId = uint64(0)
)

func NextIqId() string {
	iqId += 1
	ret := strconv.FormatUint(iqId, 10)
	return ret
}

func (g *GatewayInfo) ReceivedXMPP_Presence(presence *xmpp.Presence) {
	if presence.Type == Type_error {
		return
	}

	transfertPresence := false

	jid := strings.SplitN(presence.From, "/", 2)
	steamJid := strings.SplitN(strings.SplitN(presence.To, "/", 2)[0], "@", 2)
	if len(jid) == 2 {
		// Resource exist —> client speaking
		if presence.Type == Type_available {
			g.XMPP_Connected_Client[presence.From] = true
		} else if presence.Type == Type_unavailable {
			delete(g.XMPP_Connected_Client, presence.From)
		}
	}

	if presence.Type == Type_probe {
		steamFriendStatus := g.FriendSteamId[steamJid[0]]
		if steamFriendStatus != nil {
			g.SendXmppPresence(steamFriendStatus.XMPP_Status, steamFriendStatus.XMPP_Type, "", steamJid[0]+"@"+XmppJidComponent, steamFriendStatus.SteamGameName, steamFriendStatus.SteamName)
		}

	} else if presence.Type == Type_subscribe {
		// Send presence to tell that the JID has been added to roster
		g.SendXmppPresence("", Type_subscribed, presence.From, presence.To, g.XMPP_JID_Client, "")

	} else if presence.Type == Type_subscribed {
	} else if presence.Type == Type_unsubscribe {
	} else if presence.Type == Type_unsubscribed {
	} else if presence.To == XmppJidComponent {
		// Destination is gateway itself
		if presence.Type == Type_unavailable {
			// Disconnect
			if len(g.XMPP_Connected_Client) <= 0 {
				g.Disconnect()
			}
		} else {
			go g.SteamConnect()
			transfertPresence = true
		}
	}

	if transfertPresence {
		// Transfert presence to Steam network
		var steamStatus steamlang.EPersonaState

		switch presence.Show {
		case Status_online:
			steamStatus = State_Online

		case Status_away:
			steamStatus = State_Away

		case Status_chat:
			steamStatus = State_Online

		case Status_extended_away:
			steamStatus = State_Snooze

		case Status_do_not_disturb:
			steamStatus = State_Busy
		}

		if g.IsSteamConnected() {
			g.SendSteamPresence(steamStatus)
			g.SendXmppPresence(presence.Show, presence.Type, "", "", presence.Status, "")
		}
	}
}

func (g *GatewayInfo) ReceivedXMPP_Message(message *xmpp.Message) {
	steamID := strings.SplitN(message.To, "@", 2)[0]
	if message.Composing != nil {
		g.SendSteamMessageComposing(steamID)
	} else if message.Paused != nil {
		return
	} else if message.Inactive != nil {
		return
	} else if message.Gone != nil {
		g.SendSteamMessageLeaveConversation(steamID)
	} else {
		if message.Body != "" {
			g.SendSteamMessage(steamID, message.Body)
		}
	}
}

func (g *GatewayInfo) ReceivedXMPP_IQ(iq *xmpp.Iq) bool {
	ret := false
	if g.XMPP_IQ_RemoteRoster_Request[iq.Id] == RemoteRosterRequestPermission {
		delete(g.XMPP_IQ_RemoteRoster_Request, iq.Id)

		if iq.Type == xmpp.IQTypeError && iq.Error.Condition() == xmpp.ErrorForbidden {
			g.AllowEditRoster = false
			logger.Info.Printf("Set allow roster edition to %v", g.AllowEditRoster)
		} else if iq.Type == xmpp.IQTypeSet && iq.PayloadName().Space == xmpp.NSRemoteRosterManager {
			remoteRosterQuery := &xmpp.RemoteRosterManagerQuery{}
			iq.PayloadDecode(remoteRosterQuery)
			if remoteRosterQuery.Type == xmpp.RemoteRosterManagerTypeAllowed {
				g.AllowEditRoster = true
			} else if remoteRosterQuery.Type == xmpp.RemoteRosterManagerTypeRejected {
				g.AllowEditRoster = false
			} else {
				g.AllowEditRoster = false
			}
			logger.Info.Printf("Set allow roster edition to %v", g.AllowEditRoster)
			g.SendXmppMessage(XmppJidComponent, "", "Set allow roster edition to "+strconv.FormatBool(g.AllowEditRoster))
		} else {
			logger.Info.Printf("Check roster edition authorisation by querying roster's user")
			// Remote roster namespace may not be supported (like prosody), so we send a roster query
			iqId := NextIqId()
			g.XMPP_IQ_RemoteRoster_Request[iqId] = RemoteRosterRequestRoster
			iqSend := &xmpp.Iq{Id: iqId, Type: xmpp.IQTypeGet, From: iq.To, To: iq.From}
			iqSend.PayloadEncode(&xmpp.RosterQuery{})
			g.XMPP_Out <- iqSend
		}
		ret = true

	} else if g.XMPP_IQ_RemoteRoster_Request[iq.Id] == RemoteRosterRequestRoster {
		delete(g.XMPP_IQ_RemoteRoster_Request, iq.Id)
		if iq.Type == xmpp.IQTypeResult && iq.PayloadName().Space == xmpp.NSRoster {
			g.AllowEditRoster = true
		} else {
			g.AllowEditRoster = false
		}
		logger.Info.Printf("Set allow roster edition to %v", g.AllowEditRoster)
		g.SendXmppMessage(XmppJidComponent, "", "Set allow roster edition to "+strconv.FormatBool(g.AllowEditRoster))
		ret = true
	}

	return ret
}

func (g *GatewayInfo) XMPP_Disconnect() {
	g.SendXmppPresence(Status_offline, Type_unavailable, "", "", "", "")
}

func (g *GatewayInfo) SendXmppPresence(status, tpye, to, from, message, nick string) {
	p := xmpp.Presence{}

	if status != "" {
		p.Show = status
	}
	if tpye != "" {
		p.Type = tpye
	}
	if message != "" {
		p.Status = message
	}
	if nick != "" {
		p.Nick = nick
	}
	if to == "" {
		p.To = g.XMPP_JID_Client
	} else {
		p.To = to
	}
	if from == "" {
		p.From = XmppJidComponent + "/" + resource
	} else {
		p.From = from + "/" + resource
	}

	if tpye == Type_subscribe && g.AllowEditRoster {
		g.addUserIntoRoster(from, nick)
	} else {
		logger.Info.Printf("Send presence %v", p)
		g.XMPP_Out <- p
	}
}

func (g *GatewayInfo) addUserIntoRoster(jid, nick string) {
	iq := xmpp.Iq{To: g.XMPP_JID_Client, Type: xmpp.IQTypeSet, Id: NextIqId()}
	query := &xmpp.RosterQuery{}
	queryItem := &xmpp.RosterItem{JID: jid, Name: nick, Subscription: xmpp.RosterSubscriptionBoth}
	queryItem.Groupes = append(queryItem.Groupes, XmppGroupUser)
	query.Items = append(query.Items, *queryItem)
	iq.PayloadEncode(query)
	logger.Info.Printf("Add user into roster %v", iq)
	g.XMPP_Out <- iq
}

func (g *GatewayInfo) removeAllUserFromRoster() {
	// Friends
	for steamId, _ := range g.SteamClient.Social.Friends.GetCopy() {
		iq := xmpp.Iq{To: g.XMPP_JID_Client, Type: xmpp.IQTypeSet, Id: NextIqId()}
		query := &xmpp.RosterQuery{}
		query.Items = append(query.Items, *&xmpp.RosterItem{JID: steamId.ToString() + "@" + XmppJidComponent, Subscription: xmpp.RosterSubscriptionRemove})
		iq.PayloadEncode(query)
		logger.Info.Printf("Remove steam user roster")
		g.XMPP_Out <- iq
	}
	// Myself
	iq := xmpp.Iq{To: g.XMPP_JID_Client, Type: xmpp.IQTypeSet, Id: NextIqId()}
	query := &xmpp.RosterQuery{}
	query.Items = append(query.Items, *&xmpp.RosterItem{JID: g.SteamClient.SteamId().ToString() + "@" + XmppJidComponent, Subscription: xmpp.RosterSubscriptionRemove})
	iq.PayloadEncode(query)
	logger.Info.Printf("Remove steam user roster")
	g.XMPP_Out <- iq
}

func (g *GatewayInfo) SendXmppMessage(from, subject, message string) {
	g.sendXmppMessage(from, subject, message, &xmpp.Active{})
	return

	g.stopComposingTimer(from)

	// Make inactive after 2 min if nothing happen
	t := time.AfterFunc(120*time.Second, func() {
		g.sendXmppMessage(from, "", "", &xmpp.Inactive{})
	})
	g.XMPP_Composing_Timers[from] = t
}

func (g *GatewayInfo) SendXmppMessageLeaveConversation(from string) {
	g.sendXmppMessage(from, "", "", &xmpp.Gone{})
	return

	g.stopComposingTimer(from)
}

func (g *GatewayInfo) SendXmppMessageComposing(from string) {
	g.sendXmppMessage(from, "", "", &xmpp.Composing{})
	return

	g.stopComposingTimer(from)

	timer := time.AfterFunc(20*time.Second, func() {
		g.sendXmppMessage(from, "", "", &xmpp.Paused{})

		t := time.AfterFunc(100*time.Second, func() {
			g.sendXmppMessage(from, "", "", &xmpp.Inactive{})
		})
		g.XMPP_Composing_Timers[from] = t
	})
	g.XMPP_Composing_Timers[from] = timer
}

func (g *GatewayInfo) stopComposingTimer(from string) {
	if t, ok := g.XMPP_Composing_Timers[from]; ok {
		// Delete previous timer if exist
		if !t.Stop() {
			// Prevent firing after stop
			<-t.C
		}
		delete(g.XMPP_Composing_Timers, from)
	}
}

func (g *GatewayInfo) sendXmppMessage(from, subject, message string, chatState interface{}) {
	if from != XmppJidComponent || from == XmppJidComponent && g.DebugMessage {
		m := xmpp.Message{To: g.XMPP_JID_Client, From: from, Body: message, Type: "chat"}

		if subject != "" {
			m.Subject = subject
		}

		switch v := chatState.(type) {
		case *xmpp.Active:
			m.Active = v
		case *xmpp.Composing:
			m.Composing = v
		case *xmpp.Paused:
			m.Paused = v
		case *xmpp.Inactive:
			m.Inactive = v
		case *xmpp.Gone:
			m.Gone = v
		default:
			m.Active = &xmpp.Active{}
		}

		logger.Info.Printf("Send message %v", m)
		g.XMPP_Out <- m
	}
}
