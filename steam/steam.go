package steam

import (
	"github.com/Philipp15b/go-steam"
	"github.com/Philipp15b/go-steam/internal/steamlang"
	"github.com/Philipp15b/go-steam/steamid"

	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"
	"log"
)

const (
	sentryFile = "sentry"
	serverAddrs = "servers.addr"

	State_Offline = steamlang.EPersonaState_Offline
	State_Online = steamlang.EPersonaState_Online
	State_Busy = steamlang.EPersonaState_Busy
	State_Away = steamlang.EPersonaState_Away
	State_Snooze = steamlang.EPersonaState_Snooze
	State_LookingToTrade = steamlang.EPersonaState_LookingToTrade
	State_LookingToPlay = steamlang.EPersonaState_LookingToPlay
	State_Max = steamlang.EPersonaState_Max

	ActionConnected = "steam_connected"
	ActionDisconnected = "steam_disconnected"

	LogInfo = "\t[STEAM INFO]\t"
	LogError = "\t[STEAM ERROR]\t"
	LogDebug = "\t[STEAM DEBUG]\t"
)

var (
  Username = ""
  Password = ""
  AuthCode = ""

  myLoginInfo = new(steam.LogOnDetails)
	client = steam.NewClient()

	ChanPresence = make(chan string)
	ChanPresenceSteam = make(chan steamlang.EPersonaState)
  ChanMessage = make(chan string)
  ChanAction = make(chan string)
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
				ChanAction <- ActionDisconnected
				// Re run Steam
				go func() {
					time.Sleep(2 * time.Second)
					Run()
				}()
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

	// TODO think again
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
	// TODO XMPP notification ofline from all steam friend

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



//------------------------------------------------------------------------------
// First authentification error
//------------------------------------------------------------------------------
// 2015/10/17 15:15:29 Steam connected
// SteamID:  STEAM_0:0:0
// 2015/10/17 15:15:30 Steam connected
// FatalErrorEvent
// Login error: EResult_AccountLogonDenied
// &{}
//------------------------------------------------------------------------------
// Authentificated
//------------------------------------------------------------------------------
// &{1593367805 SsEgwfLu7Spol4jEzA0}
// &{}
// &{EClientPersonaStateFlag_GameDataBlob | EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_LastSeen | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_QueryPort | EClientPersonaStateFlag_SourceID | EClientPersonaStateFlag_Status STEAM_0:0:6208314 EPersonaState_Online 0 0 0  0 0 65535 STEAM_0:0:0 [] Chteufleur eb267fc00a974dbd2ae7e3c3de06130729cdfcbd 1445087561 1445088090 0  1 1 false  0}
// &{EClientPersonaStateFlag_GameDataBlob | EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_LastSeen | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_QueryPort | EClientPersonaStateFlag_SourceID | EClientPersonaStateFlag_Status STEAM_0:1:23476646 EPersonaState_Online 0 274190 274190 Broforce 0 0 0 STEAM_0:0:0 [] moumoutte 1cebd6410e6ceb02af60d42a34ca355c5b3b41e6 1445034635 1445034642 0  1 1 false  0}
// &{EClientPersonaStateFlag_GameDataBlob | EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_LastSeen | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_QueryPort | EClientPersonaStateFlag_SourceID | EClientPersonaStateFlag_Status STEAM_0:1:22486691 EPersonaState_Online 0 0 0  0 0 65535 STEAM_0:0:0 [] Pimouss 9ed44c8628c419cb9cb184014889063de4146e38 1445078411 1445086045 0  1 1 false  0}
// &{EClientPersonaStateFlag_GameDataBlob | EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_LastSeen | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_QueryPort | EClientPersonaStateFlag_SourceID | EClientPersonaStateFlag_Status STEAM_0:1:20592317 EPersonaState_Online 0 0 0  0 0 65535 STEAM_0:0:0 [] Francis LALAN 9fe3eb4c1b2abf6768df2d41c6343dc69ec27e83 1445042279 1445072749 0  1 1 false  0}
// &{103582791434277245 0 EAccountFlags_PersonaNameSet | EAccountFlags_Unbannable Steam Trading Cards Group 2ce81e0b8f1c748f86d1ca4230a7f45dd0b906b1 1448289 367203 23 124404 [] []}
// &{EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:0:6208314 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] Chteufleur eb267fc00a974dbd2ae7e3c3de06130729cdfcbd 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:0:6208314 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] Chteufleur eb267fc00a974dbd2ae7e3c3de06130729cdfcbd 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:0:52837827 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] beutard 0000000000000000000000000000000000000000 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID 103582791434277245 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] Steam Trading Cards Group 2ce81e0b8f1c748f86d1ca4230a7f45dd0b906b1 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:1:20592317 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] Francis LALAN 9fe3eb4c1b2abf6768df2d41c6343dc69ec27e83 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:1:23476646 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] moumoutte 1cebd6410e6ceb02af60d42a34ca355c5b3b41e6 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:1:94714518 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] mimiptitsourie fa0f46837771ec60665889c285f9496d353a50e6 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:0:18218832 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] Milamber 3cfcc21ea34862a3ee9634d36c91391cf73b4af9 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:0:61725424 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] azer9911 c160d49f7a842f408051bbada040b7d154bbcaf5 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:1:17516002 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] Earenfly 52d70ad89eab9bf4f64f25509ac80a8081b0c0d6 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:1:20054215 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] Wurrzounet 257038af8628ded724c8c8af072c678aadf3c72c 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:0:45567500 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] johann.gautier.fr bcc9b911486843390af223b8e7cb6d931e9b7e75 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:1:22486691 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] Pimouss 9ed44c8628c419cb9cb184014889063de4146e38 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:1:45984569 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] Sl@ny aa4bf32f5fc568c7949bf88c3c569c452edfcdf6 0 0 0  0 0 false  0}
// &{EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:1:19128186 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] pouzetalexis 0000000000000000000000000000000000000000 0 0 0  0 0 false  0}
// &{Chteufleur SE [] [] 38 false EAccountFlags_EmailValidated | EAccountFlags_HWIDSet | EAccountFlags_LogonExtraSecurity | EAccountFlags_PasswordSet | EAccountFlags_PersonaNameSet | EAccountFlags_Steam2MigrationComplete 0 }
// &{EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_SourceID STEAM_0:0:6208314 EPersonaState_Offline 0 0 0  0 0 0 STEAM_0:0:0 [] Chteufleur eb267fc00a974dbd2ae7e3c3de06130729cdfcbd 0 0 0  0 0 false  0}
//
//
//
//
//
// &{STEAM_0:0:0 STEAM_0:1:22486691  EChatEntryType_Typing}
// &{STEAM_0:0:0 STEAM_0:1:22486691 bouh EChatEntryType_ChatMsg}
// &{EClientPersonaStateFlag_GameDataBlob | EClientPersonaStateFlag_GameExtraInfo | EClientPersonaStateFlag_LastSeen | EClientPersonaStateFlag_PlayerName | EClientPersonaStateFlag_Presence | EClientPersonaStateFlag_QueryPort | EClientPersonaStateFlag_SourceID | EClientPersonaStateFlag_Status STEAM_0:1:23476646 EPersonaState_Away 0 274190 274190 Broforce 0 0 0 STEAM_0:0:0 [] moumoutte 1cebd6410e6ceb02af60d42a34ca355c5b3b41e6 1445034635 1445034642 0  1 1 false  0}
//------------------------------------------------------------------------------
// Servers list
//------------------------------------------------------------------------------
// 2015/10/18 11:43:49 	[STEAM DEBUG]	%!(EXTRA *steam.ClientCMListEvent=&{[185.25.180.14:27019 155.133.242.8:27019 185.25.180.15:27019 155.133.242.9:27017 185.25.180.15:27017 185.25.180.14:27018 155.133.242.9:27020 185.25.180.14:27017 185.25.180.15:27018 155.133.242.9:27019 155.133.242.8:27018 162.254.197.41:27020 162.254.197.41:27021 162.254.197.41:27017 162.254.197.42:27018 162.254.197.40:27019 185.25.180.14:27020 185.25.180.15:27020 162.254.197.42:27020 162.254.197.41:27018 155.133.242.9:27018 162.254.197.42:27017 162.254.197.42:27021 155.133.242.8:27020 162.254.197.42:27019 162.254.197.40:27017 162.254.197.40:27021 162.254.197.40:27020 155.133.242.8:27017 162.254.197.41:27019 162.254.197.40:27018 162.254.196.42:27019 162.254.196.42:27021 162.254.196.40:27017 162.254.196.42:27017 162.254.196.43:27021 162.254.196.41:27019 162.254.196.43:27017 162.254.196.41:27020 162.254.196.41:27021 162.254.196.43:27020 162.254.196.42:27018 162.254.196.41:27018 162.254.196.43:27019 162.254.196.42:27020 162.254.196.40:27018 162.254.196.40:27021 162.254.196.40:27020 162.254.196.43:27018 162.254.196.40:27019 162.254.196.41:27017 146.66.152.10:27019 146.66.152.11:27018 146.66.152.10:27018 146.66.152.11:27019 146.66.152.11:27017 146.66.152.10:27017 146.66.152.10:27020 146.66.152.11:27020 185.25.182.10:27019 185.25.182.10:27020 146.66.155.8:27019 185.25.182.10:27017 185.25.182.10:27018 146.66.155.8:27017 146.66.155.8:27020 146.66.155.8:27018 208.78.164.13:27018 208.78.164.10:27018 208.78.164.12:27018 208.78.164.10:27019 208.78.164.12:27019 208.78.164.9:27019 208.78.164.11:27019 208.78.164.14:27018 208.78.164.13:27019 208.78.164.11:27018 208.78.164.9:27018 208.78.164.13:27017 208.78.164.10:27017]})
