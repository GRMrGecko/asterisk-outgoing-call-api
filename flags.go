package main

import (
	"flag"
	"fmt"
	"os"
)

// Flags are command line tick options.
type Flags struct {
	ConfigPath string
	HTTPBind   string
	HTTPPort   uint
}

// Init configures the golang flags and parses the command line provided options.
func (f *Flags) Init() {
	flag.Usage = func() {
		fmt.Printf("asterisk-outgoing-call-api: Make an outgoing call via an API call.\n\nUsage:\n")
		flag.PrintDefaults()
	}

	var printVersion bool
	flag.BoolVar(&printVersion, "v", false, "Print version")

	var usage string
	usage = "Load configuration from file."
	flag.StringVar(&f.ConfigPath, "config", "", usage)
	flag.StringVar(&f.ConfigPath, "c", "", usage+" (shorthand)")

	usage = "Bind address for http server"
	flag.StringVar(&f.HTTPBind, "http-bind", "", usage)
	flag.StringVar(&f.HTTPBind, "b", "", usage+" (shorthand)")

	usage = "Bind port for http server"
	flag.UintVar(&f.HTTPPort, "http-port", 0, usage)
	flag.UintVar(&f.HTTPPort, "p", 0, usage+" (shorthand)")

	flag.Parse()

	if printVersion {
		fmt.Println("asterisk-outgoing-call-api: 0.1")
		os.Exit(0)
	}
}
