package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/getlantern/systray"
	"github.com/kbinani/screenshot"
	"github.com/martinlindhe/inputbox"
	_ "github.com/qodrorid/godaemon"
	"github.com/shibukawa/configdir"
	"github.com/sqweek/dialog"
)

//go:generate go run internal/resources.go

type state int

//Config of the app
//stored in:
// =windows> %LOCALAPPDATA% (C:\\Users\\<User>\\AppData\\Local)
// =macos>	 ${HOME}/Library/Caches
// =linux>	 ${XDG_CACHE_HOME} (${HOME}/.cache)
type Config struct {
	IP      string
	Port    uint16
	Display int
}

//Connection string
func (c Config) Connection() string {
	ipString := c.IP
	if c.Port != 0 && c.Port != defaultConfig.Port {
		ipString += fmt.Sprintf("%d", c.Port)
	}
	return ipString
}

var defaultConfig = Config{"127.0.0.1", 56565, 0}

const (
	appName = "GoScreen Client"
	//Version of the application
	Version           = "1.0.0"
	stateNoServer     = state(0)
	stateConnected    = state(1)
	stateDisconnected = state(2)
)

var startStreamingMenuItem *systray.MenuItem
var stopStreamingMenuItem *systray.MenuItem
var displayMenuItems []*systray.MenuItem
var stop = func() {}

var config Config

func main() {
	loadConfig()
	systray.Run(onReady, onExit)
}

func loadConfig() {
	configDirs := configdir.New("ryupold", appName)
	configDirs.LocalPath, _ = filepath.Abs(".")
	folder := configDirs.QueryFolderContainsFile("settings.json")
	if folder != nil {
		data, _ := folder.ReadFile("settings.json")
		json.Unmarshal(data, &config)
	} else {
		config = defaultConfig
	}
}

func saveConfig() error {
	data, err := json.Marshal(&config)
	if err != nil {
		return err
	}

	configDirs := configdir.New("ryupold", appName)
	folders := configDirs.QueryFolders(configdir.All)
	return folders[0].WriteFile("settings.json", data)
}

func start() {
	ctx, cancel := context.WithCancel(context.Background())
	stop = cancel
	err := startStreaming(ctx, config.IP, config.Port)
	if err != nil {
		log(err)
		setStreamingState(stateDisconnected)
		systray.SetTooltip(err.Error())
	}
}

func updateCheckboxes() {
	for i, item := range displayMenuItems {
		// x := "  "
		if config.Display == i {
			// x = "X"
			item.Check()
		} else {
			item.Uncheck()
		}
		// item.SetTitle(fmt.Sprintf("[%s] Display #%d", x, i+1))

	}
}

func onReady() {
	if runtime.GOOS == "darwin" {
		systray.SetTitle("")
	} else {
		systray.SetTitle(appName)
	}

	//menu items
	connectMenuItem := systray.AddMenuItem("Connect...", "Connect to a streaming target")
	systray.AddSeparator()
	startStreamingMenuItem = systray.AddMenuItem("Start streaming", "Start streaming to ")
	stopStreamingMenuItem = systray.AddMenuItem("Stop streaming", "stop streaming your screen")
	systray.AddSeparator()
	for i := 0; i < screenshot.NumActiveDisplays(); i++ {
		displayItem := systray.AddMenuItem(fmt.Sprintf("Display #%d", i+1), "")
		displayMenuItems = append(displayMenuItems, displayItem)
		index := i
		go func() {
			for {
				_, ok := <-displayItem.ClickedCh
				if ok {
					config.Display = index
					saveConfig()
					updateCheckboxes()
				}
			}
		}()
	}
	updateCheckboxes()
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
				ip, ok := inputbox.InputBox("Connect...", fmt.Sprintf("Enter IP:Port (default: %d)", defaultConfig.Port), config.Connection())
				if ok && strings.TrimSpace(ip) != "" {
					fmt.Println("you entered:", ip)
					parts := strings.Split(strings.TrimSpace(ip), ":")
					if len(parts) == 1 {
						config = Config{parts[0], defaultConfig.Port, config.Display}
						if err := saveConfig(); err != nil {
							dialog.Message("error while saving config: %s", err.Error()).Title("Error").Error()
							setStreamingState(stateNoServer)
						} else {
							setStreamingState(stateDisconnected)
						}
					} else if len(parts) == 2 {
						port, err := strconv.Atoi(parts[1])
						if err != nil {
							dialog.Message("Invalid port. Remove :#### to use the default port %d", defaultConfig.Port).Title("Error").Error()
							setStreamingState(stateNoServer)
						} else {
							config = Config{parts[0], uint16(port), config.Display}
							if err = saveConfig(); err != nil {
								dialog.Message("error while saving config: %s", err.Error()).Title("Error").Error()
								setStreamingState(stateNoServer)
							} else {
								setStreamingState(stateDisconnected)
							}
						}
					} else {
						dialog.Message("should look more like this -> %s:%d", defaultConfig.IP, defaultConfig.Port).Title("WTF òÓ").Error()
						setStreamingState(stateNoServer)
					}
				} else {
					fmt.Println("cancelled")
				}
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
				dialog.Message("%s %s", appName, Version).Title(appName).Info()
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
		systray.SetIcon(binICOIconNoServer)
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
