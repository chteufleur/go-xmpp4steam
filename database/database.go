package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

const (
	databaseFile = "go_xmpp4steam.db"

	createDatabaseStmt = "create table if not exists users (jid text not null primary key, steamLogin text, steamPwd text, debug int);"
	insertDatabaseStmt = "insert into users (jid, steamLogin, steamPwd, debug) values(?, ?, ?, ?)"
	deleteDatabaseStmt = "delete from users where jid=?"
	selectDatabaseStmt = "select jid, steamLogin, steamPwd, debug from users where jid=?"
	updateDatabaseStmt = "update users set steamLogin=?, steamPwd=?, debug=? where jid=?"

	LogInfo  = "\t[SQLITE INFO]\t"
	LogError = "\t[SQLITE ERROR]\t"
	LogDebug = "\t[SQLITE DEBUG]\t"
)

type DatabaseLine struct {
	Jid        string
	SteamLogin string
	SteamPwd   string
	Debug      bool
}

var (
	db = new(sql.DB)
)

func init() {
	d, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		log.Printf("%sError on openning database", LogError, err)
	}
	db = d

	_, err = db.Exec(createDatabaseStmt)
	if err != nil {
		log.Printf("%sFailed to create table", LogError, err)
	}
}

func Close() {
	db.Close()
}

func (newLine *DatabaseLine) AddLine() bool {
	log.Printf("%sAdd new line %v", LogInfo, newLine)

	isUserRegistred := getLine(newLine.Jid) != nil
	if isUserRegistred {
		return newLine.UpdateLine()
	}

	stmt, err := db.Prepare(insertDatabaseStmt)
	if err != nil {
		log.Printf("%sError on insert jid %s", LogError, newLine.Jid, err)
		return false
	}
	defer stmt.Close()
	debug := 0
	if newLine.Debug {
		debug = 1
	}
	_, err = stmt.Exec(newLine.Jid, newLine.SteamLogin, newLine.SteamPwd, debug)
	if err != nil {
		log.Printf("%sError on creating SQL statement", LogError, err)
		return false
	}

	return true
}

func (newLine *DatabaseLine) UpdateLine() bool {
	log.Printf("%sUpdate line %s", LogInfo, newLine.Jid)
	stmt, err := db.Prepare(updateDatabaseStmt)
	if err != nil {
		log.Printf("%sError on update ", LogError, err)
		return false
	}
	defer stmt.Close()
	debug := 0
	if newLine.Debug {
		debug = 1
	}
	_, err = stmt.Exec(newLine.SteamLogin, newLine.SteamPwd, debug, newLine.Jid)
	if err != nil {
		log.Printf("%sError on updating SQL statement", LogError, err)
		return false
	}

	return true
}

func RemoveLine(jid string) bool {
	// Update Steam login and password to blank before deleting,
	// because it is not really deleted in the SQLite file.
	line := new(DatabaseLine)
	line.Jid = jid
	line.UpdateLine()

	log.Printf("%sRemove line %s", LogInfo, jid)
	stmt, err := db.Prepare(deleteDatabaseStmt)
	if err != nil {
		log.Printf("%sError on delete jid %s", LogError, jid, err)
		return false
	}
	defer stmt.Close()
	res, err := stmt.Exec(jid)
	if err != nil {
		log.Printf("%sError on delete SQL statement", LogError, err)
		return false
	}

	affect, err := res.RowsAffected()
	if err != nil {
		log.Printf("%sError on delete SQL statement", LogError, err)
		return false
	}
	if affect == 0 {
		log.Printf("%sNo line affected", LogDebug)
		return false
	}

	return true
}

func GetLine(jid string) *DatabaseLine {
	ret := getLine(jid)

	if ret == nil || ret.SteamLogin == "" {
		log.Printf("%sLine empty", LogDebug)
		return nil
	}

	return ret
}

func getLine(jid string) *DatabaseLine {
	log.Printf("%sGet line %s", LogInfo, jid)
	ret := new(DatabaseLine)

	stmt, err := db.Prepare(selectDatabaseStmt)
	if err != nil {
		log.Printf("%sError on select line", LogError, err)
		return nil
	}
	defer stmt.Close()
	debug := 0
	err = stmt.QueryRow(jid).Scan(&ret.Jid, &ret.SteamLogin, &ret.SteamPwd, &debug)
	if err != nil {
		log.Printf("%sError on select scan", LogError, err)
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
	log.Printf("%sGet all lines", LogInfo)
	var ret []DatabaseLine

	rows, err := db.Query("select jid, steamLogin, steamPwd from users")
	if err != nil {
		log.Printf("%sError on select query", LogError, err)
	}
	defer rows.Close()
	for rows.Next() {
		user := new(DatabaseLine)
		rows.Scan(&user.Jid, &user.SteamLogin, &user.SteamPwd)
		if user.SteamLogin != "" {
			ret = append(ret, *user)
		}
	}

	return ret
}
