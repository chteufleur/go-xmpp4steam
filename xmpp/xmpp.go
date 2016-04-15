package xmpp

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp.git"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/gateway"

	"log"
	"strings"
)

const (
	ActionConnexion       = "action_xmpp_connexion"
	ActionDeconnexion     = "action_xmpp_deconnexion"
	ActionMainMethodEnded = "action_xmpp_main_method_ended"

	LogInfo  = "\t[XMPP COMPONENT INFO]\t"
	LogError = "\t[XMPP COMPONENT ERROR]\t"
	LogDebug = "\t[XMPP COMPONENT DEBUG]\t"
)

var (
	Addr   = "127.0.0.1:5347"
	JidStr = ""
	Secret = ""

	jid    xmpp.JID
	stream = new(xmpp.Stream)
	comp   = new(xmpp.XMPP)

	ChanAction = make(chan string)

	Debug = true

	MapGatewayInfo = make(map[string]*gateway.GatewayInfo)
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
	// Define xmpp out for all users
	for _, u := range MapGatewayInfo {
		u.XMPP_Out = comp.Out
	}

	for x := range comp.In {
		switch v := x.(type) {
		case *xmpp.Presence:
			jidBare := strings.SplitN(v.From, "/", 2)[0]
			g := MapGatewayInfo[jidBare]
			if g != nil {
				log.Printf("%sPresence transfered to %s", LogDebug, jidBare)
				g.ReceivedXMPP_Presence(v)
			} else {
				if v.Type != gateway.Type_error && v.Type != gateway.Type_probe {
					SendPresence(gateway.Status_offline, gateway.Type_unavailable, jid.Domain, v.From, "Your are not registred", "")
				}
			}

		case *xmpp.Message:
			jidBare := strings.SplitN(v.From, "/", 2)[0]
			g := MapGatewayInfo[jidBare]
			if g != nil {
				log.Printf("%sMessage transfered to %s", LogDebug, jidBare)
				g.ReceivedXMPP_Message(v)
			} else {
				SendMessage(v.From, "", "Your are not registred. If you want to register, please, send an Ad-Hoc command.")
			}

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
	SendPresence(gateway.Status_offline, gateway.Type_unavailable, "", "", "", "")
}

func must(v interface{}, err error) interface{} {
	if err != nil {
		log.Fatal(LogError, err)
	}
	return v
}

func Disconnect() {
	log.Printf("%sXMPP disconnect", LogInfo)
	SendPresence(gateway.Status_offline, gateway.Type_unavailable, "", "", "", "")
}

func SendPresence(status, tpye, from, to, message, nick string) {
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
	if from == "" {
		p.From = jid.Domain
	} else {
		p.From = from
	}
	if to != "" {
		p.To = to
	}

	comp.Out <- p
}

func SendMessage(to, subject, message string) {
	m := xmpp.Message{From: jid.Domain, To: to, Body: message, Type: "chat"}

	if subject != "" {
		m.Subject = subject
	}

	log.Printf("%sSenp message %v", LogInfo, m)
	comp.Out <- m
}

func AddNewUser(jid, steamLogin, steamPwd string) {
	log.Printf("%sAdd user %s to the map", LogInfo, jid)

	// TODO Move Gateway creation into right package
	g := new(gateway.GatewayInfo)
	g.SteamLogin = steamLogin
	g.SteamPassword = steamPwd
	g.XMPP_JID_Client = jid
	g.SentryFile = gateway.SentryDirectory + jid
	g.FriendSteamId = make(map[string]struct{})

	MapGatewayInfo[jid] = g
	go g.Run()
}
