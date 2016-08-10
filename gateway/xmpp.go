package gateway

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp.git/src/xmpp"
	"github.com/Philipp15b/go-steam/protocol/steamlang"

	"log"
	"strings"
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

	LogXmppInfo  = "\t[XMPP INFO]\t"
	LogXmppError = "\t[XMPP ERROR]\t"
	LogXmppDebug = "\t[XMPP DEBUG]\t"
)

var (
	XmppJidComponent = ""
)

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
		} else if presence.Type == Type_available {
			go g.SteamConnect()
			transfertPresence = true
		}

	} else {
		// Destination is Steam user
		if presence.Type == Type_unavailable {
			// Disconnect
			if len(g.XMPP_Connected_Client) <= 0 {
				g.Disconnect()
			}
		} else if presence.Type == Type_available {
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
			g.SendXmppPresence(presence.Show, presence.Type, presence.From, "", presence.Status, "")
		}
	}
}

func (g *GatewayInfo) ReceivedXMPP_Message(message *xmpp.Message) {
	steamID := strings.SplitN(message.To, "@", 2)[0]
	g.SendSteamMessage(steamID, message.Subject+"\n"+message.Body)
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
		// TODO add an option to allow message comming directly from the gateway
		p.From = XmppJidComponent + "/" + resource
	} else {
		p.From = from + "/" + resource
	}

	log.Printf("%sSend presence %v", LogXmppInfo, p)
	g.XMPP_Out <- p
}

func (g *GatewayInfo) SendXmppMessage(from, subject, message string) {
	if from != XmppJidComponent || from == XmppJidComponent && g.DebugMessage {
		m := xmpp.Message{To: g.XMPP_JID_Client, From: from, Body: message, Type: "chat"}

		if subject != "" {
			m.Subject = subject
		}

		log.Printf("%sSend message %v", LogXmppInfo, m)
		g.XMPP_Out <- m
	}
}
