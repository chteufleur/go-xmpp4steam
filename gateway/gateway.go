package gateway

import (
	"github.com/Philipp15b/go-steam"

	"time"
)

const (
	resource = "go-xmpp4steam"
)

var (
	SentryDirectory = "sentries/"
	XmppGroupUser   = "Steam"

	RemoteRosterRequestPermission = "remote-roster-request-permission"
	RemoteRosterRequestRoster     = "remote-roster-request-roster"
)

type GatewayInfo struct {
	// Steam
	SteamLogin      string
	SteamPassword   string
	SteamLoginInfo  *steam.LogOnDetails
	SteamClient     *steam.Client
	SentryFile      string
	FriendSteamId   map[string]*StatusSteamFriend
	SteamConnecting bool
	Deleting        bool

	// XMPP
	XMPP_JID_Client              string
	XMPP_Out                     chan interface{}
	XMPP_Connected_Client        map[string]bool
	XMPP_Composing_Timers        map[string]*time.Timer
	DebugMessage                 bool
	XMPP_IQ_RemoteRoster_Request map[string]string
	AllowEditRoster              bool
}

type StatusSteamFriend struct {
	XMPP_Status   string
	XMPP_Type     string
	SteamGameName string
	SteamName     string
}

func (g *GatewayInfo) Run() {
	go g.SteamRun()
}

func (g *GatewayInfo) SetSteamAuthCode(authCode string) {
	g.SteamLoginInfo.AuthCode = authCode
}

func (g *GatewayInfo) Disconnect() {
	go g.XMPP_Disconnect()
	go g.SteamDisconnect()
	g.SteamConnecting = false
}

func (g *GatewayInfo) Delete() {
	g.Deleting = true

	if g.AllowEditRoster {
		g.removeAllUserFromRoster()
	}

	g.Disconnect()
}
