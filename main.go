package main

import (
	"context"
	"flag"
	"os"

	ttserver "github.com/tristan-weil/ttserver/server"
	tthandler "github.com/tristan-weil/ttserver/server/handler"
	ttfinger "github.com/tristan-weil/ttserver/server/handler/finger"
	ttgopher "github.com/tristan-weil/ttserver/server/handler/gopher"
	ttutils "github.com/tristan-weil/ttserver/utils"
)

func main() {
	//
	// parameters
	//
	configFilePtr := flag.String("config", "", ""+
		"The path of the config file")
	helpPtr := flag.Bool("help", false, ""+
		"Display the help message.")
	versionPtr := flag.Bool("version", false, ""+
		"Display the version.")
	flag.Parse()

	configFile := *configFilePtr
	version := *versionPtr
	help := *helpPtr

	//
	// main
	//
	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if version {
		ttutils.PrintVersion()
		os.Exit(0)
	}

	serveConnHandlers := map[string]tthandler.IServeConnHandler{
		"finger": new(ttfinger.Handler),
		"gopher": new(ttgopher.Handler),
	}

	// start
	var ctx = context.Background()

	s := ttserver.NewManager(ctx, serveConnHandlers, configFile)

	if err := s.Start(); err != nil {
		s.Logger().Error(err)
		s.Logger().Fatal("exiting !")
	} else {
		s.Logger().Infof("exiting !")
	}
}
