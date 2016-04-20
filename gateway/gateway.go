package gateway

import (
	"github.com/Philipp15b/go-steam"
)

const (
	SentryDirectory = "sentries/"
)

type GatewayInfo struct {
	// Steam
	SteamLogin      string
	SteamPassword   string
	SteamLoginInfo  *steam.LogOnDetails
	SteamClient     *steam.Client
	SentryFile      string
	FriendSteamId   map[string]struct{}
	SteamConnecting bool

	// XMPP
	XMPP_JID_Client       string
	XMPP_Out              chan interface{}
	XMPP_Connected_Client map[string]bool
}

func (g *GatewayInfo) Run() {
	go g.SteamRun()
}

func (g *GatewayInfo) SetSteamAuthCode(authCode string) {
	g.SteamLoginInfo.AuthCode = authCode
}
