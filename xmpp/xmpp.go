package xmpp

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp.git/src/xmpp"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/database"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/gateway"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/logger"

	"os"
	"strings"
	"time"
)

const (
	ActionConnexion       = "action_xmpp_connexion"
	ActionDeconnexion     = "action_xmpp_deconnexion"
	ActionMainMethodEnded = "action_xmpp_main_method_ended"
)

var (
	Addr   = "127.0.0.1:5347"
	JidStr = ""
	Secret = ""

	SoftVersion = ""

	jid    xmpp.JID
	stream = new(xmpp.Stream)
	comp   = new(xmpp.XMPP)

	ChanAction = make(chan string)

	Debug = true

	MapGatewayInfo = make(map[string]*gateway.GatewayInfo)
	AdminUsers     = make(map[string]bool)
	startTime      = time.Now()
)

func Run() {
	logger.Info.Printf("Running")
	// Create stream and configure it as a component connection.
	jid = must(xmpp.ParseJID(JidStr)).(xmpp.JID)
	stream = must(xmpp.NewStream(Addr, &xmpp.StreamConfig{LogStanzas: Debug})).(*xmpp.Stream)
	comp = must(xmpp.NewComponentXMPP(stream, jid, Secret)).(*xmpp.XMPP)

	mainXMPP()
	time.Sleep(1 * time.Second)
	logger.Info.Printf("Reach XMPP Run method's end")
	go Run()
}

func mainXMPP() {
	defer logger.Info.Printf("Reach main method's end")
	// Define xmpp out for all users
	for _, u := range MapGatewayInfo {
		u.XMPP_Out = comp.Out
	}

	for x := range comp.In {
		switch v := x.(type) {
		case *xmpp.Presence:
			jidBareFrom := strings.SplitN(v.From, "/", 2)[0]
			jidBareTo := strings.SplitN(v.To, "/", 2)[0]
			g := MapGatewayInfo[jidBareFrom]
			if g != nil {
				if jidBareTo == jid.Domain || v.Type == gateway.Type_probe {
					// Forward only if presence is for component or is type probe, in order not to spam set presence on Steam
					logger.Debug.Printf("Presence transferred to %s", jidBareFrom)
					go g.ReceivedXMPP_Presence(v)
				}
			} else {
				if v.Type != gateway.Type_error && v.Type != gateway.Type_probe {
					SendPresence(gateway.Status_offline, gateway.Type_unavailable, jid.Domain, v.From, "Your are not registred", "")
				}
			}

		case *xmpp.Message:
			jidBareFrom := strings.SplitN(v.From, "/", 2)[0]
			g := MapGatewayInfo[jidBareFrom]
			if g != nil {
				logger.Debug.Printf("Message transferred to %s", jidBareFrom)
				go g.ReceivedXMPP_Message(v)
			} else {
				SendMessage(v.From, "", "Your are not registred. If you want to register, please, send an Ad-Hoc command.")
			}

		case *xmpp.Iq:
			jidBareFrom := strings.SplitN(v.From, "/", 2)[0]
			jidBareTo := strings.SplitN(v.To, "/", 2)[0]

			g := MapGatewayInfo[jidBareFrom]
			iqTreated := false
			if g != nil {
				logger.Debug.Printf("Iq transferred to %s", jidBareFrom)
				iqTreated = g.ReceivedXMPP_IQ(v)
			}

			if !iqTreated {
				switch v.PayloadName().Space {
				case xmpp.NSDiscoInfo:
					execDisco(v)

				case xmpp.NSDiscoItems:
					execDisco(v)

				case xmpp.NodeAdHocCommand:
					if jidBareTo == jid.Domain {
						execCommandAdHoc(v)
					} else {
						sendNotSupportedFeature(v)
					}

				case xmpp.NSVCardTemp:
					if jidBareTo == jid.Domain {
						reply := v.Response(xmpp.IQTypeResult)
						vcard := &xmpp.VCard{}
						reply.PayloadEncode(vcard)
						comp.Out <- reply
					} else {
						sendNotSupportedFeature(v)
					}

				case xmpp.NSJabberClient:
					if jidBareTo == jid.Domain {
						reply := v.Response(xmpp.IQTypeResult)
						reply.PayloadEncode(&xmpp.SoftwareVersion{Name: "go-xmpp4steam", Version: SoftVersion})
						comp.Out <- reply
					} else {
						sendNotSupportedFeature(v)
					}

				case xmpp.NSRegister:
					if jidBareTo == jid.Domain {
						treatmentNSRegister(v)
					} else {
						sendNotSupportedFeature(v)
					}

				case xmpp.NSRoster:
					// Do nothing

				case xmpp.NSPing:
					if jidBareTo == jid.Domain {
						treatmentNSPing(v)
					} else {
						sendNotSupportedFeature(v)
					}

				default:
					sendNotSupportedFeature(v)
				}
			}

		default:
			logger.Debug.Printf("recv: %v", x)
		}
	}
}

func must(v interface{}, err error) interface{} {
	if err != nil {
		logger.Debug.Printf("%v", err)
		os.Exit(1)
	}
	return v
}

func treatmentNSRegister(iq *xmpp.Iq) {
	reply := iq.Response(xmpp.IQTypeResult)
	jidBareFrom := strings.SplitN(iq.From, "/", 2)[0]
	registerQuery := &xmpp.RegisterQuery{}

	if iq.Type == xmpp.IQTypeGet {
		registerQuery.Instructions = "Please provide your Steam login and password (Please, be aware that the given Steam account information will be saved into an un-encrypted SQLite database)."

		dbUser := database.GetLine(jidBareFrom)
		if dbUser != nil {
			// User already registered
			registerQuery.Registered = &xmpp.RegisterRegistered{}
			registerQuery.Username = dbUser.SteamLogin
			registerQuery.XForm = *getXFormRegistration(dbUser.SteamLogin)
		} else {
			registerQuery.XForm = *getXFormRegistration("")
		}
		reply.PayloadEncode(registerQuery)

	} else if iq.Type == xmpp.IQTypeSet {
		iq.PayloadDecode(registerQuery)

		if registerQuery.Remove != nil {
			RemoveUser(jidBareFrom)
		} else {
			dbUser := getUser(registerQuery.XForm.Fields, iq)
			if dbUser != nil {
				if dbUser.UpdateUser() {
					AddNewUser(dbUser.Jid, dbUser.SteamLogin, dbUser.SteamPwd, dbUser.Debug)
				} else {
					reply.Type = xmpp.IQTypeError
					reply.Error = xmpp.NewErrorWithCode("406", "modify", xmpp.ErrorNotAcceptable, "")
				}
			} else {
				reply.Type = xmpp.IQTypeError
				reply.Error = xmpp.NewErrorWithCode("409", "cancel", xmpp.ErrorConflict, "")
			}
		}
	}
	comp.Out <- reply
}

func treatmentNSPing(iq *xmpp.Iq) {
	reply := iq.Response(xmpp.IQTypeResult)
	comp.Out <- reply
}

func sendNotSupportedFeature(iq *xmpp.Iq) {
	if iq.Type != xmpp.IQTypeError && iq.Type != xmpp.IQTypeResult {
		reply := iq.Response(xmpp.IQTypeError)
		reply.PayloadEncode(xmpp.NewError("cancel", xmpp.ErrorFeatureNotImplemented, ""))
		comp.Out <- reply
	}
}

func Disconnect() {
	logger.Info.Printf("XMPP disconnect")
	for _, u := range MapGatewayInfo {
		u.SteamDisconnect()
	}
	comp.Close()
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
	m := xmpp.Message{From: jid.Domain, To: to, Type: "chat"}
	mBody := xmpp.MessageBody{Value: message}
	m.Body = append(m.Body, mBody)

	if subject != "" {
		m.Subject = subject
	}

	logger.Info.Printf("Senp message %v", m)
	comp.Out <- m
}

func AddNewUser(jidUser, steamLogin, steamPwd string, debugMessage bool) {
	logger.Info.Printf("Add user %s to the map (debug mode set to %v)", jidUser, debugMessage)

	g := new(gateway.GatewayInfo)
	g.SteamLogin = steamLogin
	g.SteamPassword = steamPwd
	g.XMPP_JID_Client = jidUser
	g.SentryFile = gateway.SentryDirectory + jidUser
	g.FriendSteamId = make(map[string]*gateway.StatusSteamFriend)
	g.Deleting = false

	g.XMPP_Out = comp.Out
	g.XMPP_Connected_Client = make(map[string]bool)
	g.DebugMessage = debugMessage
	g.XMPP_IQ_RemoteRoster_Request = make(map[string]string)
	g.AllowEditRoster = false

	MapGatewayInfo[jidUser] = g
	go g.Run()

	logger.Info.Printf("Check roster edition by asking with remote roster manager namespace")
	// Ask if remote roster is allow
	iqId := gateway.NextIqId()
	g.XMPP_IQ_RemoteRoster_Request[iqId] = gateway.RemoteRosterRequestPermission
	iq := xmpp.Iq{To: jidUser, From: jid.Domain, Type: xmpp.IQTypeGet, Id: iqId}
	// iq.PayloadEncode(&xmpp.RosterQuery{})
	iq.PayloadEncode(&xmpp.RemoteRosterManagerQuery{Reason: "Manage contacts in the Steam contact list", Type: xmpp.RemoteRosterManagerTypeRequest})
	comp.Out <- iq
}

func RemoveUser(jidBare string) bool {
	ret := database.RemoveLine(jidBare)

	if ret {
		g := MapGatewayInfo[jidBare]
		ret = g != nil
		if ret {
			g.Delete()
			MapGatewayInfo[jidBare] = nil
		}
	}

	return ret
}
