package xmpp

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp.git"

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

	LogInfo  = "\t[XMPP INFO]\t"
	LogError = "\t[XMPP ERROR]\t"
	LogDebug = "\t[XMPP DEBUG]\t"
)

var (
	Addr   = "127.0.0.1:5347"
	JidStr = ""
	Secret = ""

	PreferedJID = ""

	jid    xmpp.JID
	stream = new(xmpp.Stream)
	comp   = new(xmpp.XMPP)

	ChanPresence = make(chan string)
	ChanMessage  = make(chan string)
	ChanAction   = make(chan string)

	CurrentStatus = Status_offline

	setJIDconnected = make(map[string]bool)

	Debug = true
)

func Run() {
	log.Printf("%sRunning", LogInfo)
	// Create stream and configure it as a component connection.
	jid = must(xmpp.ParseJID(JidStr)).(xmpp.JID)
	stream = must(xmpp.NewStream(Addr, &xmpp.StreamConfig{LogStanzas: Debug})).(*xmpp.Stream)
	comp = must(xmpp.NewComponentXMPP(stream, jid, Secret)).(*xmpp.XMPP)

	mainXMPP()
	ChanAction <- ActionMainMethodEnded
}

func mainXMPP() {
	for x := range comp.In {
		switch v := x.(type) {
		case *xmpp.Presence:
			if strings.SplitN(v.From, "/", 2)[0] == PreferedJID && v.Type != Type_probe {
				if v.Type == Type_unavailable && v.To == JidStr {
					delete(setJIDconnected, v.From)
					log.Printf("%sPresence reçut unavailable", LogDebug)
					if len(setJIDconnected) <= 0 {
						// Disconnect only when all JID are disconnected
						Disconnect()
						ChanAction <- ActionDeconnexion
					}
				} else if v.Type == Type_subscribe {
					SendPresenceFrom("", Type_subscribed, JidStr, "", "")
				} else if v.Type != Type_subscribed { // Type subscribed is send by JID without ressourse
					log.Printf("%sAdd connected user. JID : %s", LogInfo, v.From)
					setJIDconnected[v.From] = true
					CurrentStatus = v.Show
					ChanAction <- ActionConnexion
				}

				ChanPresence <- v.Show
				ChanPresence <- v.Type
			}

		case *xmpp.Message:
			steamID := strings.SplitN(v.To, "@", 2)[0]
			ChanMessage <- steamID
			ChanMessage <- v.Body

		case *xmpp.Iq:
			switch v.PayloadName().Space {
			case xmpp.NsDiscoItems:
				execDiscoCommand(v)

			case xmpp.NodeAdHocCommand:
				execCommandAdHoc(v)
			}

		default:
			log.Printf("%srecv: %v", LogDebug, x)
		}
	}

	// Send deconnexion
	SendPresence(Status_offline, Type_unavailable, "")
}

func must(v interface{}, err error) interface{} {
	if err != nil {
		log.Fatal(LogError, err)
	}
	return v
}

func Disconnect() {
	log.Printf("%sXMPP disconnect", LogInfo)
	SendPresence(Status_offline, Type_unavailable, "")
}

func SendPresence(status, tpye, message string) {
	comp.Out <- xmpp.Presence{To: PreferedJID, From: jid.Domain, Show: status, Type: tpye, Status: message}
}

func SendPresenceFrom(status, tpye, from, message, nick string) {
	/*
		if message == "" {
			comp.Out <- xmpp.Presence{To: PreferedJID, From: from, Show: status, Type: tpye, Nick: nick}
		} else if nick == "" {
			comp.Out <- xmpp.Presence{To: PreferedJID, From: from, Show: status, Type: tpye, Status: message}
		} else if nick == "" {
			comp.Out <- xmpp.Presence{To: PreferedJID, From: from, Show: status, Type: tpye}
		} else {
			comp.Out <- xmpp.Presence{To: PreferedJID, From: from, Show: status, Type: tpye, Status: message, Nick: nick}
		}
	*/
	p := xmpp.Presence{To: PreferedJID, From: from}

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

	comp.Out <- p
}

func SendMessage(from, message string) {
	comp.Out <- xmpp.Message{To: PreferedJID, From: from, Body: message}
}
