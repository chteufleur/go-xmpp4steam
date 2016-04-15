package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

const (
	databaseFile = "go_xmpp4steam.db"

	createDatabaseStmt = "create table if not exists users (jid text not null primary key, steamLogin text, steamPwd text);"
	insertDatabaseStmt = "insert into users (jid, steamLogin, steamPwd) values(?, ?, ?)"
	deleteDatabaseStmt = "delete from users where jid = ?"
	selectDatabaseStmt = "select jid, steamLogin, steamPwd from users where jid = ?"

	LogInfo  = "\t[SQLITE INFO]\t"
	LogError = "\t[SQLITE ERROR]\t"
	LogDebug = "\t[SQLITE DEBUG]\t"
)

type DatabaseLine struct {
	Jid        string
	SteamLogin string
	SteamPwd   string
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
	stmt, err := db.Prepare(insertDatabaseStmt)
	if err != nil {
		log.Printf("%sError on insert jid %s", LogError, newLine.Jid, err)
		return false
	}
	defer stmt.Close()
	_, err = stmt.Exec(newLine.Jid, newLine.SteamLogin, newLine.SteamPwd)
	if err != nil {
		log.Printf("%sError on creating SQL statement", LogError, err)
		return false
	}

	return true
}

func RemoveLine(jid string) bool {
	// FIXME not working
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
		return false
	}

	return true
}

func GetLine(jid string) *DatabaseLine {
	log.Printf("%sGet line %s", LogInfo, jid)
	ret := new(DatabaseLine)

	stmt, err := db.Prepare(selectDatabaseStmt)
	if err != nil {
		log.Printf("%sError on select line", LogError, err)
		return nil
	}
	defer stmt.Close()

	err = stmt.QueryRow(jid).Scan(&ret.Jid, &ret.SteamLogin, &ret.SteamPwd)
	if err != nil {
		log.Printf("%sError on select scan", LogError, err)
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
		ret = append(ret, *user)
	}

	return ret
}
