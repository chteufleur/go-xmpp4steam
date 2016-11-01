package database

import (
	"database/sql"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/logger"
	_ "github.com/mattn/go-sqlite3"
)

const (
	DatabaseFileName = "go_xmpp4steam.db"

	createDatabaseStmt    = "create table if not exists users (jid text not null primary key, steamLogin text, steamPwd text, debug int);"
	insertDatabaseStmt    = "insert into users (jid, steamLogin, steamPwd, debug) values(?, ?, ?, ?)"
	deleteDatabaseStmt    = "delete from users where jid=?"
	selectDatabaseStmt    = "select jid, steamLogin, steamPwd, debug from users where jid=?"
	selectAllDatabaseStmt = "select jid, steamLogin, steamPwd, debug from users"
	updateDatabaseStmt    = "update users set steamLogin=?, steamPwd=?, debug=? where jid=?"
)

type DatabaseLine struct {
	Jid        string
	SteamLogin string
	SteamPwd   string
	Debug      bool
}

var (
	db           = new(sql.DB)
	DatabaseFile = ""
)

func init() {
}

func Init() {
	logger.Info.Printf("Init database (file %s)", DatabaseFile)
	d, err := sql.Open("sqlite3", DatabaseFile)
	if err != nil {
		logger.Error.Printf("Error on openning database : %v", err)
	}
	db = d

	_, err = db.Exec(createDatabaseStmt)
	if err != nil {
		logger.Error.Printf("Failed to create table : %v", err)
	}
}

func Close() {
	db.Close()
}

func (newLine *DatabaseLine) AddLine() bool {
	logger.Info.Printf("Add new line %v", newLine)

	isUserRegistred := getLine(newLine.Jid) != nil
	if isUserRegistred {
		return newLine.UpdateLine()
	}

	stmt, err := db.Prepare(insertDatabaseStmt)
	if err != nil {
		logger.Error.Printf("Error on insert jid %s : %v", newLine.Jid, err)
		return false
	}
	defer stmt.Close()
	debug := 0
	if newLine.Debug {
		debug = 1
	}
	_, err = stmt.Exec(newLine.Jid, newLine.SteamLogin, newLine.SteamPwd, debug)
	if err != nil {
		logger.Error.Printf("Error on creating SQL statement : %v", err)
		return false
	}

	return true
}

func (newLine *DatabaseLine) UpdateLine() bool {
	logger.Info.Printf("Update line %s", newLine.Jid)
	stmt, err := db.Prepare(updateDatabaseStmt)
	if err != nil {
		logger.Error.Printf("Error on update : %v", err)
		return false
	}
	defer stmt.Close()
	debug := 0
	if newLine.Debug {
		debug = 1
	}
	if newLine.SteamPwd == "" {
		oldLine := GetLine(newLine.Jid)
		newLine.SteamPwd = oldLine.SteamPwd
	}
	_, err = stmt.Exec(newLine.SteamLogin, newLine.SteamPwd, debug, newLine.Jid)
	if err != nil {
		logger.Error.Printf("Error on updating SQL statement : %v", err)
		return false
	}

	return true
}

func (dbUser *DatabaseLine) UpdateUser() bool {
	isUserRegistred := GetLine(dbUser.Jid) != nil
	var isSqlSuccess bool
	if isUserRegistred {
		isSqlSuccess = dbUser.UpdateLine()
	} else {
		isSqlSuccess = dbUser.AddLine()
	}
	return isSqlSuccess
}

func RemoveLine(jid string) bool {
	// Update Steam login and password to blank before deleting,
	// because it is not really deleted in the SQLite file.
	line := new(DatabaseLine)
	line.Jid = jid
	line.UpdateLine()

	logger.Info.Printf("Remove line %s", jid)
	stmt, err := db.Prepare(deleteDatabaseStmt)
	if err != nil {
		logger.Error.Printf("Error on delete jid %s : %v", jid, err)
		return false
	}
	defer stmt.Close()
	res, err := stmt.Exec(jid)
	if err != nil {
		logger.Error.Printf("Error on delete SQL statement : %v", err)
		return false
	}

	affect, err := res.RowsAffected()
	if err != nil {
		logger.Error.Printf("Error on delete SQL statement : %v", err)
		return false
	}
	if affect == 0 {
		logger.Debug.Printf("No line affected")
		return false
	}

	return true
}

func GetLine(jid string) *DatabaseLine {
	ret := getLine(jid)

	if ret == nil || ret.SteamLogin == "" {
		logger.Debug.Printf("Line empty")
		return nil
	}

	return ret
}

func getLine(jid string) *DatabaseLine {
	logger.Info.Printf("Get line %s", jid)
	ret := new(DatabaseLine)

	stmt, err := db.Prepare(selectDatabaseStmt)
	if err != nil {
		logger.Error.Printf("Error on select line : %v", err)
		return nil
	}
	defer stmt.Close()
	debug := 0
	err = stmt.QueryRow(jid).Scan(&ret.Jid, &ret.SteamLogin, &ret.SteamPwd, &debug)
	if err != nil {
		logger.Error.Printf("Error on select scan : %v", err)
		return nil
	}
	if debug == 1 {
		ret.Debug = true
	} else {
		ret.Debug = false
	}

	return ret
}

func GetAllLines() []DatabaseLine {
	logger.Info.Printf("Get all lines")
	var ret []DatabaseLine

	rows, err := db.Query(selectAllDatabaseStmt)
	if err != nil {
		logger.Error.Printf("Error on select query : %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		user := new(DatabaseLine)
		debug := 0
		rows.Scan(&user.Jid, &user.SteamLogin, &user.SteamPwd, &debug)
		if user.SteamLogin != "" {
			if debug == 1 {
				user.Debug = true
			} else {
				user.Debug = false
			}
			ret = append(ret, *user)
		}
	}

	return ret
}
