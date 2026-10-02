package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vault-thirteen/File-Uploader/pkg/settings"
	"github.com/vault-thirteen/File-Uploader/pkg/uploader"

	ver "github.com/vault-thirteen/auxie/Versioneer/classes/Versioneer"
)

const (
	OsExitCode_NormalShutdown   = 0
	OsExitCode_BadArgsCount     = 1
	OsExitCode_BadShutdown      = 2
	OsExitCode_ServerHasCrashed = 3
	OsExitCode_Anomaly          = 4
)

const (
	Usage = "Usage:  [tool.exe] <Settings_File>"
)

const (
	OsSignalIsReceived = "OS signal is received:"
	ServerHasCrashed   = "server has crashed"
)

func main() {
	showIntro()

	args, err := NewArgumentsFromOs()
	if err != nil {
		fmt.Println(Usage)
		os.Exit(OsExitCode_BadArgsCount)
		return
	}

	var s *settings.Settings
	s, err = settings.NewSettingsFromFile(args.SettingsFilePath)
	mustBeNoError(err)

	var u *uploader.Uploader
	u, err = uploader.NewUploader(s)
	mustBeNoError(err)

	err = u.Start()
	mustBeNoError(err)

	exitCode := work(u)
	os.Exit(exitCode)
}

func mustBeNoError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func showIntro() {
	versioneer, err := ver.New(false)
	mustBeNoError(err)
	versioneer.ShowIntroText("Server")
	versioneer.ShowComponentsInfoText()
	fmt.Println()
}

func work(u *uploader.Uploader) (exitCode int) {
	var isRunning bool
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case osSignal := <-osSignals:
			{
				log.Println(OsSignalIsReceived, osSignal)

				err := u.Stop()
				if err != nil {
					log.Println(err)
					return OsExitCode_BadShutdown
				}

				return OsExitCode_NormalShutdown
			}

		default:
			{
				isRunning = u.IsRunning()
				if isRunning {
					time.Sleep(time.Second)
					break
				}

				log.Println(ServerHasCrashed)
				return OsExitCode_ServerHasCrashed
			}
		} // Select.
	} // For.

	return OsExitCode_Anomaly
}
