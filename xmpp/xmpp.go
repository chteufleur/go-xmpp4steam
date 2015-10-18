package xmpp


import (
  // "github.com/emgee/go-xmpp"
  "go-xmpp"

  "log"
  "strings"
)

const (
  Status_online = ""
  Status_offline = ""
  Status_away = "away"
  Status_chat = "chat"
  Status_do_not_disturb = "dnd"
  Status_extended_away = "xa"

  Type_available = ""
  Type_unavailable = "unavailable"

  ActionConnexion = "action_xmpp_connexion"
  ActionDeconnexion = "action_xmpp_deconnexion"

  LogInfo = "\t[XMPP INFO]\t"
  LogError = "\t[XMPP ERROR]\t"
  LogDebug = "\t[XMPP DEBUG]\t"
)

var (
	Addr   = "127.0.0.1:5347"
  JidStr = ""
	Secret = ""

  PreferedJID = ""

  jid xmpp.JID
  stream = new(xmpp.Stream)
  comp   = new(xmpp.XMPP)

  ChanPresence = make(chan string)
  ChanMessage = make(chan string)
  ChanAction = make(chan string)

  CurrentStatus = Status_offline
)


func Run() {
  log.Printf("%sRunning", LogInfo)
	// Create stream and configure it as a component connection.
	jid = must(xmpp.ParseJID(JidStr)).(xmpp.JID)
	stream = must(xmpp.NewStream(Addr, &xmpp.StreamConfig{LogStanzas: true})).(*xmpp.Stream)
	comp = must(xmpp.NewComponentXMPP(stream, jid, Secret)).(*xmpp.XMPP)

  SendPresence(Status_online, Type_available)

  mainXMPP()
}

func mainXMPP() {
	for x := range comp.In {
    switch v := x.(type) {
    case *xmpp.Presence:
      if strings.SplitN(v.From, "/", 2)[0] == PreferedJID && v.To == JidStr {
        if v.Type == Type_unavailable {
          Disconnect()
          ChanAction <- ActionDeconnexion
        } else {
          // SendPresence(v.Show, v.Type)
          CurrentStatus = v.Show
          ChanAction <- ActionConnexion
        }

        ChanPresence <- v.Show
      }

    case *xmpp.Message:
      steamID := strings.SplitN(v.To, "@", 2)[0]
      ChanMessage <- steamID
      ChanMessage <- v.Body

    default:
	    log.Printf("%srecv: %v", LogDebug, x)
    }
	}

  // Send deconnexion
  SendPresence(Status_offline, Type_unavailable)
}

func must(v interface{}, err error) interface{} {
	if err != nil {
		log.Fatal(LogError, err)
	}
	return v
}

func Disconnect() {
  SendPresence(Status_offline, Type_unavailable)
}

func SendPresence(status, tpye string) {
  comp.Out <- xmpp.Presence{To: PreferedJID, From: jid.Domain, Show: status, Type: tpye}
}

func SendPresenceFrom(status, tpye, from string) {
  comp.Out <- xmpp.Presence{To: PreferedJID, From: from, Show: status, Type: tpye}
}

func SendMessage(from, message string) {
  comp.Out <- xmpp.Message{To: PreferedJID, From: from, Body: message}
}