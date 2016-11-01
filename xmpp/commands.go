package xmpp

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp.git/src/xmpp"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/database"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/logger"

	"fmt"
	"strings"
	"time"
)

const (
	CommandAuthcode           = "steamAuthCodeCommand"
	CommandGetIdentifiants    = "steamGetIdentifiants"
	CommandDisconnectSteam    = "disconnectSteam"
	CommandRemoveRegistration = "removeRegistration"
	CommandToggleDebugMode    = "toggleDebugMode"
	CommandUptimeMode         = "uptime"
	CommandMessageBroadcast   = "messageBroadcast"
)

var (
	ChanAuthCode    = make(chan string)
	identityGateway = &xmpp.DiscoIdentity{Category: "gateway", Type: "steam", Name: "Steam Gateway"}
	identityClients = &xmpp.DiscoIdentity{Category: "client", Type: "pc", Name: "Steam client"}
)

func execDiscoCommand(iq *xmpp.Iq) {
	logger.Info.Printf("Ad-Hoc Command")

	// Disco Ad-Hoc
	reply := iq.Response(xmpp.IQTypeResult)
	discoItem := &xmpp.DiscoItems{Node: xmpp.NodeAdHocCommand}

	jidBareFrom := strings.SplitN(iq.From, "/", 2)[0]
	jidBareTo := strings.SplitN(iq.To, "/", 2)[0]
	dbUser := database.GetLine(jidBareFrom)

	if jidBareTo == jid.Domain {
		// Ad-Hoc command only on gateway
		// Add available commands
		if dbUser == nil {
			discoI := &xmpp.DiscoItem{JID: jid.Domain, Node: CommandGetIdentifiants, Name: "Steam registration"}
			discoItem.Item = append(discoItem.Item, *discoI)
		} else {
			// Add only if user is registered
			discoI := &xmpp.DiscoItem{JID: jid.Domain, Node: CommandAuthcode, Name: "Add Steam Auth Code"}
			discoItem.Item = append(discoItem.Item, *discoI)
			discoI = &xmpp.DiscoItem{JID: jid.Domain, Node: CommandDisconnectSteam, Name: "Force Steam deconnexion"}
			discoItem.Item = append(discoItem.Item, *discoI)
			discoI = &xmpp.DiscoItem{JID: jid.Domain, Node: CommandRemoveRegistration, Name: "Remove registration"}
			discoItem.Item = append(discoItem.Item, *discoI)
			discoI = &xmpp.DiscoItem{JID: jid.Domain, Node: CommandToggleDebugMode, Name: "Toggle debug mode"}
			discoItem.Item = append(discoItem.Item, *discoI)
			discoI = &xmpp.DiscoItem{JID: jid.Domain, Node: CommandUptimeMode, Name: "Uptime"}
			discoItem.Item = append(discoItem.Item, *discoI)
		}

		if AdminUsers[jidBareFrom] {
			discoI := &xmpp.DiscoItem{JID: jid.Domain, Node: CommandMessageBroadcast, Name: "Broadcast a message"}
			discoItem.Item = append(discoItem.Item, *discoI)
		}
	}

	reply.PayloadEncode(discoItem)
	comp.Out <- reply
}

func execDisco(iq *xmpp.Iq) {
	logger.Info.Printf("Disco Feature")
	jidBareTo := strings.SplitN(iq.To, "/", 2)[0]

	discoInfoReceived := &xmpp.DiscoItems{}
	iq.PayloadDecode(discoInfoReceived)

	switch iq.PayloadName().Space {
	case xmpp.NSDiscoInfo:
		reply := iq.Response(xmpp.IQTypeResult)

		discoInfo := &xmpp.DiscoInfo{}
		if jidBareTo == jid.Domain {
			// Only gateway
			discoInfo.Identity = append(discoInfo.Identity, *identityGateway)
			discoInfo.Feature = append(discoInfo.Feature, xmpp.DiscoFeature{Var: xmpp.NodeAdHocCommand})
			discoInfo.Feature = append(discoInfo.Feature, xmpp.DiscoFeature{Var: xmpp.NSJabberClient})
			discoInfo.Feature = append(discoInfo.Feature, xmpp.DiscoFeature{Var: xmpp.NSRegister})
			discoInfo.Feature = append(discoInfo.Feature, xmpp.DiscoFeature{Var: xmpp.NSPing})
		} else {
			// Only steam users
			discoInfo.Identity = append(discoInfo.Identity, *identityClients)
			discoInfo.Feature = append(discoInfo.Feature, xmpp.DiscoFeature{Var: xmpp.NSChatStatesNotification})
		}
		// Both
		discoInfo.Feature = append(discoInfo.Feature, xmpp.DiscoFeature{Var: xmpp.NSDiscoInfo})
		discoInfo.Feature = append(discoInfo.Feature, xmpp.DiscoFeature{Var: xmpp.NSDiscoItems})

		reply.PayloadEncode(discoInfo)
		comp.Out <- reply

	case xmpp.NSDiscoItems:
		if discoInfoReceived.Node == xmpp.NodeAdHocCommand {
			// Ad-Hoc command
			execDiscoCommand(iq)
		} else {
			reply := iq.Response(xmpp.IQTypeResult)
			discoItems := &xmpp.DiscoItems{}
			reply.PayloadEncode(discoItems)
			comp.Out <- reply
		}
	}
}

func execCommandAdHoc(iq *xmpp.Iq) {
	adHoc := &xmpp.AdHocCommand{}
	iq.PayloadDecode(adHoc)
	jidBareFrom := strings.SplitN(iq.From, "/", 2)[0]

	if adHoc.SessionID == "" && adHoc.Action == xmpp.ActionAdHocExecute {
		// First step in the command
		logger.Info.Printf("Ad-Hoc command (Node : %s). First step.", adHoc.Node)

		reply := iq.Response(xmpp.IQTypeResult)
		cmd := &xmpp.AdHocCommand{Node: adHoc.Node, Status: xmpp.StatusAdHocExecute, SessionID: xmpp.SessionID()}
		if adHoc.Node == CommandAuthcode {
			// Command Auth Code
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocForm, Title: "Steam Auth Code", Instructions: "Please provide the auth code that Steam sended to you."}

			field := &xmpp.AdHocField{Var: "code", Label: "Auth Code", Type: xmpp.TypeAdHocFieldTextSingle}
			cmdXForm.Fields = append(cmdXForm.Fields, *field)
			cmd.XForm = *cmdXForm

		} else if adHoc.Node == CommandGetIdentifiants {
			// Command Auth Code
			cmdXForm := getXFormRegistration("")
			cmd.XForm = *cmdXForm

		} else if adHoc.Node == CommandDisconnectSteam {
			// Command steam deconnection
			cmd.Status = xmpp.StatusAdHocCompleted
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocResult, Title: "Force Steam deconnexion"}
			cmd.XForm = *cmdXForm
			note := &xmpp.AdHocNote{Type: xmpp.TypeAdHocNoteInfo}

			g := MapGatewayInfo[jidBareFrom]
			if g != nil {
				g.Disconnect()
				note.Value = "Send deconnexion on Steam network"
			} else {
				note.Value = "Your are not registred."
			}
			cmd.Note = *note
		} else if adHoc.Node == CommandRemoveRegistration {
			// Command remove registration
			cmd.Status = xmpp.StatusAdHocCompleted
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocResult, Title: "Remove registration"}
			cmd.XForm = *cmdXForm
			note := &xmpp.AdHocNote{Type: xmpp.TypeAdHocNoteInfo}

			if RemoveUser(jidBareFrom) {
				note.Value = "Remove registration success."
			} else {
				note.Value = "Failed to remove your registration."
			}

			cmd.Note = *note
		} else if adHoc.Node == CommandToggleDebugMode {
			// Command toggle debug mode
			cmd.Status = xmpp.StatusAdHocCompleted
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocResult, Title: "Toggle debug mode"}
			cmd.XForm = *cmdXForm
			note := &xmpp.AdHocNote{Type: xmpp.TypeAdHocNoteInfo}

			dbUser := database.GetLine(jidBareFrom)
			if dbUser != nil {
				dbUser.Debug = !dbUser.Debug
				g := MapGatewayInfo[jidBareFrom]
				ok := dbUser.UpdateLine()
				if ok && g != nil {
					g.DebugMessage = dbUser.Debug
					if dbUser.Debug {
						note.Value = "Debug activated."
					} else {
						note.Value = "Debug desactivated."
					}
				} else {
					note.Value = "Failed to update your profile. :("
				}
			} else {
				note.Value = "Your not registered."
			}

			cmd.Note = *note
		} else if adHoc.Node == CommandUptimeMode {
			// Command get uptime
			cmd.Status = xmpp.StatusAdHocCompleted
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocResult, Title: "Uptime"}
			cmd.XForm = *cmdXForm
			deltaT := time.Since(startTime)
			val := fmt.Sprintf("%dj %dh %dm %ds", int64(deltaT.Hours()/24), int64(deltaT.Hours())%24, int64(deltaT.Minutes())%60, int64(deltaT.Seconds())%60)
			note := &xmpp.AdHocNote{Type: xmpp.TypeAdHocNoteInfo, Value: val}

			cmd.Note = *note
		} else if adHoc.Node == CommandMessageBroadcast && AdminUsers[jidBareFrom] {
			// Command send broadcast message
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocForm, Title: "Broadcast a message", Instructions: "Message to broadcast to all user."}

			field := &xmpp.AdHocField{Var: "message", Label: "Message", Type: xmpp.TypeAdHocFieldTextSingle}
			cmdXForm.Fields = append(cmdXForm.Fields, *field)
			cmd.XForm = *cmdXForm
		}
		reply.PayloadEncode(cmd)
		comp.Out <- reply
	} else if adHoc.Action == xmpp.ActionAdHocExecute || adHoc.Action == xmpp.ActionAdHocNext {
		// Last step in the command
		logger.Info.Printf("Ad-Hoc command (Node : %s). Last step.", adHoc.Node)
		reply := iq.Response(xmpp.IQTypeResult)
		cmd := &xmpp.AdHocCommand{Node: adHoc.Node, Status: xmpp.StatusAdHocCompleted, SessionID: adHoc.SessionID}

		if adHoc.Node == CommandAuthcode && adHoc.XForm.Type == xmpp.TypeAdHocSubmit {
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocResult, Title: "Steam Auth Code"}
			cmd.XForm = *cmdXForm
			note := &xmpp.AdHocNote{Type: xmpp.TypeAdHocNoteInfo}

			// Command Auth Code
			authCode := ""
			fields := adHoc.XForm.Fields
			for _, field := range fields {
				if field.Var == "code" {
					authCode = field.Value
					break
				}
			}
			if authCode != "" {
				// Succeeded
				g := MapGatewayInfo[jidBareFrom]
				if g != nil {
					g.SetSteamAuthCode(authCode)
					note.Value = "Command succeeded !"
				} else {
					note.Value = "Your are not registred. Please, register before sending Steam auth code."
				}
			} else {
				// Failed
				note.Value = "Error append while executing command"
			}
			cmd.Note = *note

		} else if adHoc.Node == CommandGetIdentifiants {
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocResult, Title: "Steam Account Info"}
			cmd.XForm = *cmdXForm
			note := &xmpp.AdHocNote{Type: xmpp.TypeAdHocNoteInfo}

			// Command Auth Code
			dbUser := getUser(adHoc.XForm.Fields, iq)
			if dbUser != nil {
				if dbUser.UpdateUser() {
					AddNewUser(dbUser.Jid, dbUser.SteamLogin, dbUser.SteamPwd, dbUser.Debug)
					note.Value = "Command succeeded !"
				} else {
					note.Value = "Error append while executing command"
				}
			} else {
				// Failed
				note.Value = "Failed because Steam login or Steam password is empty."
			}
			cmd.Note = *note

		} else if adHoc.Node == CommandMessageBroadcast && AdminUsers[jidBareFrom] {
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocResult, Title: "Broadcast a message"}
			cmd.XForm = *cmdXForm
			note := &xmpp.AdHocNote{Type: xmpp.TypeAdHocNoteInfo}

			// Command Auth Code
			message := ""
			fields := adHoc.XForm.Fields
			for _, field := range fields {
				if field.Var == "message" {
					message = field.Value
					break
				}
			}
			if message != "" {
				// Succeeded
				for userJID := range MapGatewayInfo {
					SendMessage(userJID, "", message)
				}
				note.Value = "Message sended to all registered users"
			} else {
				// Failed
				note.Value = "There is no message to send"
			}
			cmd.Note = *note
		}

		reply.PayloadEncode(cmd)
		comp.Out <- reply
	} else if adHoc.Action == xmpp.ActionAdHocCancel {
		// command canceled
		logger.Info.Printf("Ad-Hoc command (Node : %s). Command canceled.", adHoc.Node)
		reply := iq.Response(xmpp.IQTypeResult)
		cmd := &xmpp.AdHocCommand{Node: adHoc.Node, Status: xmpp.StatusAdHocCanceled, SessionID: adHoc.SessionID}
		reply.PayloadEncode(cmd)
		comp.Out <- reply
	}
}

func getXFormRegistration(steamLogin string) *xmpp.AdHocXForm {
	cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocForm, Title: "Steam Account Info", Instructions: "Please provide your Steam login and password (Please, be aware that the given Steam account information will be saved into an un-encrypted SQLite database)."}

	field := &xmpp.AdHocField{Var: "login", Label: "Steam Login", Type: xmpp.TypeAdHocFieldTextSingle}
	field.Value = steamLogin
	cmdXForm.Fields = append(cmdXForm.Fields, *field)
	field = &xmpp.AdHocField{Var: "password", Label: "Steam Password", Type: xmpp.TypeAdHocFieldTextPrivate}
	cmdXForm.Fields = append(cmdXForm.Fields, *field)

	return cmdXForm
}

func getUser(fields []xmpp.AdHocField, iq *xmpp.Iq) *database.DatabaseLine {
	// Command Auth Code
	steamLogin := ""
	steamPwd := ""
	for _, field := range fields {
		if field.Var == "login" {
			steamLogin = field.Value
		} else if field.Var == "password" {
			steamPwd = field.Value
		}
	}
	if steamLogin != "" {
		// Succeeded
		jidBareFrom := strings.SplitN(iq.From, "/", 2)[0]
		dbUser := new(database.DatabaseLine)
		dbUser.Jid = jidBareFrom
		dbUser.SteamLogin = steamLogin
		dbUser.SteamPwd = steamPwd
		dbUser.Debug = false

		return dbUser
	} else {
		return nil
	}
}
