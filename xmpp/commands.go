package xmpp

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp.git/src/xmpp"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/database"

	"log"
	"strings"
)

const (
	CommandAuthcode           = "steamAuthCodeCommand"
	CommandGetIdentifiants    = "steamGetIdentifiants"
	CommandDisconnectSteam    = "disconnectSteam"
	CommandRemoveRegistration = "removeRegistration"
	CommandToggleDebugMode    = "toggleDebugMode"
)

var (
	ChanAuthCode = make(chan string)
)

func execDiscoCommand(iq *xmpp.Iq) {
	log.Printf("%sDiscovery item iq received", LogInfo)
	reply := iq.Response(xmpp.IQTypeResult)
	discoItem := &xmpp.DiscoItems{Node: xmpp.NodeAdHocCommand}

	// Add available commands
	discoI := &xmpp.DiscoItem{JID: jid.Domain, Node: CommandAuthcode, Name: "Add Steam Auth Code"}
	discoItem.Item = append(discoItem.Item, *discoI)
	discoI = &xmpp.DiscoItem{JID: jid.Domain, Node: CommandGetIdentifiants, Name: "Steam registration"}
	discoItem.Item = append(discoItem.Item, *discoI)
	discoI = &xmpp.DiscoItem{JID: jid.Domain, Node: CommandDisconnectSteam, Name: "Force Steam deconnexion"}
	discoItem.Item = append(discoItem.Item, *discoI)
	discoI = &xmpp.DiscoItem{JID: jid.Domain, Node: CommandRemoveRegistration, Name: "Remove registration"}
	discoItem.Item = append(discoItem.Item, *discoI)
	discoI = &xmpp.DiscoItem{JID: jid.Domain, Node: CommandToggleDebugMode, Name: "Toggle debug mode"}
	discoItem.Item = append(discoItem.Item, *discoI)

	reply.PayloadEncode(discoItem)
	comp.Out <- reply
}

func execCommandAdHoc(iq *xmpp.Iq) {
	adHoc := &xmpp.AdHocCommand{}
	iq.PayloadDecode(adHoc)

	if adHoc.SessionID == "" && adHoc.Action == xmpp.ActionAdHocExecute {
		// First step in the command
		log.Printf("%sAd-Hoc command (Node : %s). First step.", LogInfo, adHoc.Node)

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
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocForm, Title: "Steam Account Info", Instructions: "Please provide your Steam login and password."}
			note := &xmpp.AdHocNote{Type: xmpp.TypeAdHocNoteInfo, Value: "Please, be aware that the given Steam account information will be saved into an un-encrypted SQLite database."}

			field := &xmpp.AdHocField{Var: "login", Label: "Steam Login", Type: xmpp.TypeAdHocFieldTextSingle}
			cmdXForm.Fields = append(cmdXForm.Fields, *field)
			field = &xmpp.AdHocField{Var: "password", Label: "Steam Password", Type: xmpp.TypeAdHocFieldTextPrivate}
			cmdXForm.Fields = append(cmdXForm.Fields, *field)

			cmd.XForm = *cmdXForm
			cmd.Note = *note
		} else if adHoc.Node == CommandDisconnectSteam {
			cmd.Status = xmpp.StatusAdHocCompleted
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocResult, Title: "Force Steam deconnexion"}
			cmd.XForm = *cmdXForm
			note := &xmpp.AdHocNote{Type: xmpp.TypeAdHocNoteInfo}

			jidBare := strings.SplitN(iq.From, "/", 2)[0]
			g := MapGatewayInfo[jidBare]
			if g != nil {
				g.Disconnect()
				note.Value = "Send deconnexion on Steam network"
			} else {
				note.Value = "Your are not registred."
			}
			cmd.Note = *note
		} else if adHoc.Node == CommandRemoveRegistration {
			cmd.Status = xmpp.StatusAdHocCompleted
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocResult, Title: "Remove registration"}
			cmd.XForm = *cmdXForm
			note := &xmpp.AdHocNote{Type: xmpp.TypeAdHocNoteInfo}

			jidBare := strings.SplitN(iq.From, "/", 2)[0]
			if RemoveUser(jidBare) {
				note.Value = "Remove registration success."
			} else {
				note.Value = "Failed to remove your registration."
			}

			cmd.Note = *note
		} else if adHoc.Node == CommandToggleDebugMode {
			cmd.Status = xmpp.StatusAdHocCompleted
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocResult, Title: "Toggle debug mode"}
			cmd.XForm = *cmdXForm
			note := &xmpp.AdHocNote{Type: xmpp.TypeAdHocNoteInfo}

			jidBare := strings.SplitN(iq.From, "/", 2)[0]
			dbUser := database.GetLine(jidBare)
			dbUser.Debug = !dbUser.Debug
			g := MapGatewayInfo[jidBare]
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

			cmd.Note = *note
		}
		reply.PayloadEncode(cmd)
		comp.Out <- reply
	} else if adHoc.Action == xmpp.ActionAdHocExecute {
		// Last step in the command
		log.Printf("%sAd-Hoc command (Node : %s). Last step.", LogInfo, adHoc.Node)
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
				// Succeded
				jidBare := strings.SplitN(iq.From, "/", 2)[0]
				g := MapGatewayInfo[jidBare]
				if g != nil {
					g.SetSteamAuthCode(authCode)
					note.Value = "Command succeded !"
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
			steamLogin := ""
			steamPwd := ""
			fields := adHoc.XForm.Fields
			for _, field := range fields {
				if field.Var == "login" {
					steamLogin = field.Value
				} else if field.Var == "password" {
					steamPwd = field.Value
				}
			}
			if steamLogin != "" && steamPwd != "" {
				// Succeded
				jidBare := strings.SplitN(iq.From, "/", 2)[0]
				dbUser := new(database.DatabaseLine)
				dbUser.Jid = jidBare
				dbUser.SteamLogin = steamLogin
				dbUser.SteamPwd = steamPwd
				dbUser.Debug = false

				isUserRegistred := database.GetLine(dbUser.Jid) != nil
				var isSqlSuccess bool
				if isUserRegistred {
					isSqlSuccess = dbUser.UpdateLine()
				} else {
					isSqlSuccess = dbUser.AddLine()
				}
				if isSqlSuccess {
					AddNewUser(dbUser.Jid, dbUser.SteamLogin, dbUser.SteamPwd, dbUser.Debug)
					note.Value = "Command succeded !"
				} else {
					note.Value = "Error append while executing command"
				}
			} else {
				// Failed
				note.Value = "Failed because Steam login or Steam password is empty."
			}
			cmd.Note = *note
		}

		reply.PayloadEncode(cmd)
		comp.Out <- reply
	} else if adHoc.Action == xmpp.ActionAdHocCancel {
		// command canceled
		log.Printf("%sAd-Hoc command (Node : %s). Command canceled.", LogInfo, adHoc.Node)
		reply := iq.Response(xmpp.IQTypeResult)
		cmd := &xmpp.AdHocCommand{Node: adHoc.Node, Status: xmpp.StatusAdHocCanceled, SessionID: adHoc.SessionID}
		reply.PayloadEncode(cmd)
		comp.Out <- reply
	}
}
