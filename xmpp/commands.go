package xmpp

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp.git"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/database"

	"log"
	"strings"
)

const (
	CommandAuthcode        = "steamAuthCodeCommand"
	CommandGetIdentifiants = "steamGetIdentifiants"
)

var (
	ChanAuthCode = make(chan string)
)

func execDiscoCommand(iq *xmpp.Iq) {
	log.Printf("%sDiscovery item iq received", LogInfo)
	reply := iq.Response(xmpp.IqTypeResult)
	discoItem := &xmpp.DiscoItems{Node: xmpp.NodeAdHocCommand}

	// Add available commands
	discoI := &xmpp.DiscoItem{JID: jid.Domain, Node: CommandAuthcode, Name: "Add Steam Auth Code"}
	discoItem.Item = append(discoItem.Item, *discoI)
	discoI = &xmpp.DiscoItem{JID: jid.Domain, Node: CommandGetIdentifiants, Name: "Steam registration"}
	discoItem.Item = append(discoItem.Item, *discoI)

	reply.PayloadEncode(discoItem)
	comp.Out <- reply
}

func execCommandAdHoc(iq *xmpp.Iq) {
	adHoc := &xmpp.AdHocCommand{}
	iq.PayloadDecode(adHoc)

	if adHoc.SessionId == "" && adHoc.Action == xmpp.ActionAdHocExecute {
		// First step in the command
		log.Printf("%sAd-Hoc command (Node : %s). First step.", LogInfo, adHoc.Node)

		reply := iq.Response(xmpp.IqTypeResult)
		cmd := &xmpp.AdHocCommand{Node: adHoc.Node, Status: xmpp.StatusAdHocExecute, SessionId: xmpp.SessionId()}
		if adHoc.Node == CommandAuthcode {
			// Command Auth Code
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocForm, Title: "Steam Auth Code", Instructions: "Please provide the auth code that Steam sended to you."}

			field := &xmpp.AdHocField{Var: "code", Label: "Auth Code", Type: xmpp.TypeAdHocFieldTextSingle}
			cmdXForm.Fields = append(cmdXForm.Fields, *field)
			cmd.XForm = *cmdXForm

		} else if adHoc.Node == CommandGetIdentifiants {
			// Command Auth Code
			cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocForm, Title: "Steam Account Info", Instructions: "Please provide your Steam login and password."}
			// TODO Warn that the given info is stored in database in clear
			note := &xmpp.AdHocNote{Type: xmpp.TypeAdHocNoteInfo, Value: "Please, be aware that the given Steam account information will be saved into an un-encrypted SQLite database."}

			field := &xmpp.AdHocField{Var: "login", Label: "Steam Login", Type: xmpp.TypeAdHocFieldTextSingle}
			cmdXForm.Fields = append(cmdXForm.Fields, *field)
			field = &xmpp.AdHocField{Var: "password", Label: "Steam Password", Type: xmpp.TypeAdHocFieldTextSingle}
			cmdXForm.Fields = append(cmdXForm.Fields, *field)

			cmd.XForm = *cmdXForm
			cmd.Note = *note
		}
		reply.PayloadEncode(cmd)
		comp.Out <- reply
	} else if adHoc.Action == xmpp.ActionAdHocExecute {
		// Last step in the command
		log.Printf("%sAd-Hoc command (Node : %s). Last step.", LogInfo, adHoc.Node)
		reply := iq.Response(xmpp.IqTypeResult)
		cmd := &xmpp.AdHocCommand{Node: adHoc.Node, Status: xmpp.StatusAdHocCompleted, SessionId: adHoc.SessionId}

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

				isUserRegistred := database.GetLine(dbUser.Jid) != nil
				var isSqlSuccess bool
				if isUserRegistred {
					isSqlSuccess = dbUser.UpdateLine()
				} else {
					isSqlSuccess = dbUser.AddLine()
				}
				if isSqlSuccess {
					AddNewUser(dbUser.Jid, dbUser.SteamLogin, dbUser.SteamPwd)
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
		reply := iq.Response(xmpp.IqTypeResult)
		cmd := &xmpp.AdHocCommand{Node: adHoc.Node, Status: xmpp.StatusAdHocCanceled, SessionId: adHoc.SessionId}
		reply.PayloadEncode(cmd)
		comp.Out <- reply
	}
}
