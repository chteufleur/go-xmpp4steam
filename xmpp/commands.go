package xmpp

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp.git"

	"log"
)



const (
  CommandAuthcode = "steamAuthCodeCommand"
)

var (
	ChanAuthCode   = make(chan string)
)

func execDiscoCommand(iq *xmpp.Iq) {
  log.Printf("%sDiscovery item iq received", LogInfo)
	reply := iq.Response(xmpp.IqTypeResult)
	discoItem := &xmpp.DiscoItems{Node: xmpp.NodeAdHocCommand}

  // Add available commands
	discoI := &xmpp.DiscoItem{JID: jid.Domain, Node: CommandAuthcode, Name: "Add Auth Code"}
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

    }
    reply.PayloadEncode(cmd)
    comp.Out <- reply
	} else if adHoc.Action == xmpp.ActionAdHocExecute {
		// Last step in the command
    log.Printf("%sAd-Hoc command (Node : %s). Last step.", LogInfo, adHoc.Node)
    reply := iq.Response(xmpp.IqTypeResult)
    cmd := &xmpp.AdHocCommand{Node: adHoc.Node, Status: xmpp.StatusAdHocCompleted, SessionId: adHoc.SessionId}

    if adHoc.Node == CommandAuthcode && adHoc.XForm.Type == xmpp.TypeAdHocSubmit {
      cmdXForm := &xmpp.AdHocXForm{Type: xmpp.TypeAdHocResult, Title: "Steam Auth Code "}
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
        ChanAuthCode <- authCode
        note.Value = "Commande effectuée avec succes !"
      } else {
        // Failed
        note.Value = "Une erreur c'est produite à l'exécution de la commande…"
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
