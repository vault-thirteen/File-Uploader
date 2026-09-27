package main

import (
	"fmt"
	"log"
	"os"

	"github.com/vault-thirteen/File-Uploader/pkg/settings"
	"github.com/vault-thirteen/File-Uploader/pkg/uploader"

	ver "github.com/vault-thirteen/auxie/Versioneer/classes/Versioneer"
)

const (
	OsExitCode_BadArgsCount = 1
	Usage                   = "Usage:  [tool.exe] <Settings_File>"
)

func main() {
	showIntro()

	args, err := NewArgumentsFromOs()
	if err != nil {
		fmt.Println(Usage)
		os.Exit(OsExitCode_BadArgsCount)
	}

	var s *settings.Settings
	s, err = settings.NewSettingsFromFile(args.SettingsFilePath)
	mustBeNoError(err)

	var u *uploader.Uploader
	u, err = uploader.NewUploader(s)
	mustBeNoError(err)

	err = u.Run()
	if err != nil {
		log.Fatal(err)
	}
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
