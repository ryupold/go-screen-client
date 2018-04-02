package main

import (
	"context"
	"fmt"
	"os"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

//go:generate go run internal/resources.go

type state int

const (
	appName           = "GoScreen"
	stateNoServer     = 0
	stateConnected    = 1
	stateDisconnected = 2
)

var startStreamingMenuItem *systray.MenuItem
var stopStreamingMenuItem *systray.MenuItem
var stop = func() {}

func main() {
	systray.Run(onReady, onExit)
}

func start() {
	ctx, cancel := context.WithCancel(context.Background())
	stop = cancel
	err := startStreaming(ctx, "localhost", 56565)
	if err != nil {
		log(err)
		setStreamingState(stateDisconnected)
		systray.SetTooltip(err.Error())
	}
}

func onReady() {
	systray.SetTitle(appName)

	//menu items
	connectMenuItem := systray.AddMenuItem("Connect...", "Connect to a streaming target")
	systray.AddSeparator()
	startStreamingMenuItem = systray.AddMenuItem("Start streaming", "Start streaming to ")
	stopStreamingMenuItem = systray.AddMenuItem("Stop streaming", "stop streaming your screen")
	systray.AddSeparator()
	aboutMenuItem := systray.AddMenuItem("About", "")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			case <-connectMenuItem.ClickedCh:
				dialog.Message("%s", "Do you want to continue?").Title("Are you sure?")
			case <-startStreamingMenuItem.ClickedCh:
				log("start clicked")
				setStreamingState(stateConnected)
				go start()
			case <-stopStreamingMenuItem.ClickedCh:
				log("stop clicked")
				stop()
				setStreamingState(stateDisconnected)
				systray.SetTooltip("stopped streaming")
			case <-aboutMenuItem.ClickedCh:
				log("about clicked")
				dialog.Message("%s", "Hello").Title(appName)
			}
		}
	}()

	setStreamingState(stateNoServer)
}

func setStreamingState(streamingState state) {
	if streamingState == stateConnected {
		systray.SetIcon(binICOIconOnline)
		systray.SetTooltip("you are streaming")
		startStreamingMenuItem.Disable()
		stopStreamingMenuItem.Enable()
	} else if streamingState == stateDisconnected {
		systray.SetIcon(binICOIconOffline)
		startStreamingMenuItem.Enable()
		stopStreamingMenuItem.Disable()
	} else {
		systray.SetIcon(binICOIconOffline)
		startStreamingMenuItem.Disable()
		stopStreamingMenuItem.Disable()
	}
}

func onExit() {
	// clean up here
	defer os.Exit(0)
	defer fmt.Println("Program finished")
	stop()
}
