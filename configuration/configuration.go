package configuration

import (
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/database"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/gateway"
	"git.kingpenguin.tk/chteufleur/go-xmpp4steam.git/logger"

	"github.com/jimlawless/cfg"

	"os"
	"strings"
)

const (
	XdgDirectoryName      = "xmpp4steam"
	configurationFilePath = "xmpp4steam/xmpp4steam.conf"

	PathConfEnvVariable  = "XDG_CONFIG_DIRS"
	DefaultXdgConfigDirs = "/etc/xdg"

	PathDataEnvVariable = "XDG_DATA_DIRS"
	DefaultXdgDataDirs  = "/usr/local/share/:/usr/share/"
	PreferedPathDataDir = "/usr/local/share"
)

var (
	MapConfig = make(map[string]string)
)

func Init() {
	loadConfigFile()

	dataPathDir := locateDataDirPath()
	database.DatabaseFile = dataPathDir + "/" + database.DatabaseFileName
	database.Init()

	gateway.ServerAddrs = dataPathDir + "/" + gateway.ServerAddrs
	gateway.SentryDirectory = dataPathDir + "/" + gateway.SentryDirectory
	os.MkdirAll(gateway.SentryDirectory, 0700)
}

func loadConfigFile() bool {
	ret := false
	envVariable := os.Getenv(PathConfEnvVariable)
	if envVariable == "" {
		envVariable = DefaultXdgConfigDirs
	}
	for _, path := range strings.Split(envVariable, ":") {
		logger.Debug.Println("Try to find configuration file into " + path)
		configFile := path + "/" + configurationFilePath
		if _, err := os.Stat(configFile); err == nil {
			// The config file exist
			if cfg.Load(configFile, MapConfig) == nil {
				// And has been loaded successfully
				logger.Info.Println("Find configuration file at " + configFile)
				ret = true
				break
			}
		}
	}
	return ret
}

func locateDataDirPath() string {
	ret := ""
	isDirFound := false
	envVariable := os.Getenv(PathDataEnvVariable)
	if envVariable == "" {
		envVariable = DefaultXdgDataDirs
	}
	for _, path := range strings.Split(envVariable, ":") {
		logger.Debug.Printf("Try to find data base directory into " + path)
		dbDir := path + "/" + XdgDirectoryName
		if fi, err := os.Stat(dbDir); err == nil && fi.IsDir() {
			// The database file exist
			logger.Info.Printf("Find data base directory at " + dbDir)
			isDirFound = true
			ret = dbDir
			break
		}
	}

	if !isDirFound {
		if strings.Contains(envVariable, PreferedPathDataDir) {
			ret = PreferedPathDataDir + "/" + XdgDirectoryName
		} else {
			ret = strings.Split(envVariable, ":")[0] + "/" + XdgDirectoryName
		}

		if os.MkdirAll(ret, 0700) == nil {
			logger.Info.Printf("Creating new data base directory at " + ret)
		} else {
			logger.Error.Printf("Fail to create data base directory at " + ret)
			os.Exit(1)
		}
	}
	return ret
}
