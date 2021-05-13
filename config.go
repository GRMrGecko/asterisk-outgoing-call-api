package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

// Reference from:
// https://wiki.asterisk.org/wiki/display/AST/Asterisk+Call+Files

// Config all configurations for this application.
type Config struct {
	HTTPBind              string            `json:"http_bind"`
	HTTPPort              uint              `json:"http_port"`
	HTTPDebug             bool              `json:"http_debug"`
	HTTPSystemDSocket     bool              `json:"http_systemd_socket"`
	AsteriskSpoolDir      string            `json:"asterisk_spool_dir"`
	DefaultChannel        string            `json:"default_channel"`
	DefaultCallerId       string            `json:"default_caller_id"`
	DefaultWaitTime       uint64            `json:"default_wait_time"` // 5 seconds per ring.
	DefaultMaxRetries     uint64            `json:"default_max_retries"`
	DefaultRetryTime      uint64            `json:"default_retry_time"`
	DefaultAccount        string            `json:"default_account"`
	DefaultApplication    string            `json:"default_application"`
	DefaultData           string            `json:"default_data"`
	PreventAPIApplication bool              `json:"prevent_api_application"` // For security, prevent applications from being executed via API call.
	DefaultContext        string            `json:"default_context"`
	DefaultExtension      string            `json:"default_extension"`
	DefaultPriority       string            `json:"default_priority"`
	DefaultSetVar         map[string]string `json:"default_set_var"`
	DefaultArchive        bool              `json:"default_archive"`
	APIToken              string            `json:"api_token"`
}

// ReadConfig read the configuration file into the config structure of the app.
func (a *App) ReadConfig() {
	// Get our current user for use in determining the home path.
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	// Different configuration file paths.
	localConfig, _ := filepath.Abs("./config.json")
	homeDirConfig := usr.HomeDir + "/.config/asterisk-outgoing-call-api/config.json"
	etcConfig := "/etc/asterisk/outgoing-call-api.json"

	// Store defaults first.
	app.config = Config{
		HTTPPort:              9747,
		HTTPDebug:             false,
		HTTPSystemDSocket:     false,
		AsteriskSpoolDir:      "/var/spool/asterisk",
		PreventAPIApplication: true,
		DefaultArchive:        false,
	}

	// Determine which config file to use.
	var configFile string
	if _, err := os.Stat(app.flags.ConfigPath); err == nil && app.flags.ConfigPath != "" {
		configFile = app.flags.ConfigPath
	} else if _, err := os.Stat(localConfig); err == nil {
		configFile = localConfig
	} else if _, err := os.Stat(homeDirConfig); err == nil {
		configFile = homeDirConfig
	} else if _, err := os.Stat(etcConfig); err == nil {
		configFile = etcConfig
	} else {
		log.Fatal("Unable to find a configuration file.")
	}

	// Read the config file.
	jsonFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Error reading JSON file: %v\n", err)
	}

	// Parse the config file into the configuration structure.
	err = json.Unmarshal(jsonFile, &app.config)
	if err != nil {
		log.Fatalf("Error parsing JSON file: %v\n", err)
	}
}
