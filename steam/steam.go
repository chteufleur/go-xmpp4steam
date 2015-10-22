package steam

import (
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/internal/steamlang"
	"github.com/Philipp15b/go-steam/steamid"

	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

const (
	sentryFile  = "sentry"
	serverAddrs = "servers.addr"

	State_Offline        = steamlang.EPersonaState_Offline
	State_Online         = steamlang.EPersonaState_Online
	State_Busy           = steamlang.EPersonaState_Busy
	State_Away           = steamlang.EPersonaState_Away
	State_Snooze         = steamlang.EPersonaState_Snooze
	State_LookingToTrade = steamlang.EPersonaState_LookingToTrade
	State_LookingToPlay  = steamlang.EPersonaState_LookingToPlay
	State_Max            = steamlang.EPersonaState_Max

	ActionConnected    = "steam_connected"
	ActionDisconnected = "steam_disconnected"
	ActionFatalError   = "steam_fatal_error"

	LogInfo  = "\t[STEAM INFO]\t"
	LogError = "\t[STEAM ERROR]\t"
	LogDebug = "\t[STEAM DEBUG]\t"
)

var (
	Username = ""
	Password = ""
	AuthCode = ""

	myLoginInfo = new(steam.LogOnDetails)
	client      = steam.NewClient()

	ChanPresence      = make(chan string)
	ChanPresenceSteam = make(chan steamlang.EPersonaState)
	ChanMessage       = make(chan string)
	ChanAction        = make(chan string)
)

func Run() {
	log.Printf("%sRunning", LogInfo)
	setLoginInfos()
	client = steam.NewClient()
	client.ConnectionTimeout = 10 * time.Second

	mainSteam()
}

func mainSteam() {
	for event := range client.Events() {
		switch e := event.(type) {
		case *steam.ConnectedEvent:
			client.Auth.LogOn(myLoginInfo)

		case *steam.MachineAuthUpdateEvent:
			ioutil.WriteFile(sentryFile, e.Hash, 0666)

		case *steam.LoggedOnEvent:
			SendPresence(steamlang.EPersonaState_Online)
			ChanAction <- ActionConnected

		case steam.FatalErrorEvent:
			log.Printf("%sFatalError: ", LogError, e)
			ChanAction <- ActionFatalError
			return

		case error:
			log.Printf("%s", LogError, e)

		case *steam.ClientCMListEvent:
			// Save servers addresses
			b, err := json.Marshal(*e)
			if err != nil {
				log.Printf("%sFailed to json.Marshal() servers list", LogError)
			}
			ioutil.WriteFile(serverAddrs, b, 0666)

		case *steam.PersonaStateEvent:
			// ChanPresence <- e.Name
			ChanPresence <- e.FriendId.ToString()
			ChanPresenceSteam <- e.State
			ChanPresence <- e.GameName

		case *steam.ChatMsgEvent:
			ChanMessage <- e.ChatterId.ToString()
			ChanMessage <- e.Message

		default:
			log.Printf("%s", LogDebug, e)
		}
	}
}

func setLoginInfos() {
	var sentryHash steam.SentryHash
	sentryHash, err := ioutil.ReadFile(sentryFile)

	myLoginInfo.Username = Username
	myLoginInfo.Password = Password

	if err == nil {
		myLoginInfo.SentryFileHash = sentryHash
		log.Printf("%sAuthentification by SentryFileHash", LogDebug)
	} else if AuthCode != "" {
		myLoginInfo.AuthCode = AuthCode
		log.Printf("%sAuthentification by AuthCode", LogDebug)
	} else {
		log.Printf("%sFirst authentification", LogDebug)
	}
}

func IsConnected() bool {
	return client.Connected()
}

func Connect() {
	if IsConnected() {
		log.Printf("%sTry to connect, but already connected", LogDebug)
		return
	}

	b, err := ioutil.ReadFile(serverAddrs)
	if err == nil {
		var toList steam.ClientCMListEvent
		err := json.Unmarshal(b, &toList)
		if err != nil {
			log.Printf("%sFailed to json.Unmarshal() servers list", LogError)
		} else {
			log.Printf("%sConnecting...", LogInfo)
			client.ConnectTo(toList.Addresses[0])
		}
	} else {
		log.Printf("%sFailed to read servers list file", LogError)
		client.Connect()
	}
}

func Disconnect() {
	log.Printf("%sSteam disconnect", LogInfo)
	go client.Disconnect()
}

func SendMessage(steamId, message string) {
	steamIdUint64, err := strconv.ParseUint(steamId, 10, 64)
	if err == nil {
		client.Social.SendMessage(steamid.SteamId(steamIdUint64), steamlang.EChatEntryType_ChatMsg, message)
	} else {
		log.Printf("%sFailed to get SteamId from %s", LogError, steamId)
	}
}

func SendPresence(status steamlang.EPersonaState) {
	client.Social.SetPersonaState(status)
}
