package uploader

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	cc "github.com/vault-thirteen/File-Uploader/pkg/CachedContent"
	"github.com/vault-thirteen/File-Uploader/pkg/helper"
	"github.com/vault-thirteen/File-Uploader/pkg/settings"

	ers "github.com/vault-thirteen/auxie/errors"
	af "github.com/vault-thirteen/auxie/file"
)

const (
	TmpDirPermissions                       = 0777
	NewFolderPermissions                    = 0777
	JournalFilePermissions                  = 0644
	HttpContentCacheTime                    = 60
	URL_Upload                              = "upload"
	URL_QueueSize                           = "queue"
	JournalTabulator                        = "\t"
	JournalLineEnd                          = helper.CRLF
	EnvVar_SystemTemporaryDirectory_Windows = "TMP"
	EnvVar_SystemTemporaryDirectory_Darwin  = "TMPDIR"
	EnvVar_SystemTemporaryDirectory_Linux   = "TMPDIR"
	EnvVar_SystemTemporaryDirectory_FreeBSD = "TMPDIR"
)

const (
	FormField_UserName     = "username"
	FormField_UserPassword = "userpwd"
	FormField_FolderPath   = "folderpath"
	FormField_FileHashes   = "filehashes"
	FormField_Files        = "files"
)

const (
	Err_FileHeaderIsNotAvailable = "file header is not available"
	Err_FileOperationError       = "file operation error"
	Errf_FileAlreadyExists       = `file already exists: "%s"`
	Errf_HashSumMismatch         = `hash sum mismatch, file: "%s"`
	Err_AlreadyStarted           = "already started"
	Err_AlreadyStopped           = "already stopped"
	Errf_UnknownOsType           = "unknown OS type: %s"
)

const (
	Msgf_FilesInQueue       = "File in queue: %d."
	Msgf_TemporaryDirectory = "Using temporary directory: %s"
)

type Uploader struct {
	controls      *UploaderControls
	settings      *settings.Settings
	dataFolder    string
	cachedContent *cc.CachedContent
	httpServer    *http.Server
}

func NewUploader(s *settings.Settings) (u *Uploader, err error) {
	u = &Uploader{
		controls:   NewUploaderControls(),
		settings:   s,
		dataFolder: s.DataFolder,
	}

	err = u.checkDataFolder()
	if err != nil {
		return nil, err
	}

	err = u.setTemporaryDirectory()
	if err != nil {
		return nil, err
	}

	u.cachedContent, err = cc.NewCachedContent(s.AssetsFolder, HttpContentCacheTime)
	if err != nil {
		return nil, err
	}

	u.httpServer = u.NewHttpServer(s.Host, s.Port)

	return u, nil
}
func (u *Uploader) checkDataFolder() (err error) {
	var folderExists bool
	folderExists, err = af.FolderExists(u.dataFolder)
	if err != nil {
		return err
	}

	if !folderExists {
		err = af.CreateFolderSafely(u.dataFolder, NewFolderPermissions)
		if err != nil {
			return err
		}
	}

	return nil
}
func (u *Uploader) setTemporaryDirectory() (err error) {
	if len(u.settings.TempDir) > 0 {
		err = os.MkdirAll(u.settings.TempDir, TmpDirPermissions)
		if err != nil {
			return err
		}
	}

	switch runtime.GOOS {
	case "windows": // Microsoft Windows.
		err = os.Setenv(EnvVar_SystemTemporaryDirectory_Windows, u.settings.TempDir)

	case "darwin": // Apple MacOS.
		err = os.Setenv(EnvVar_SystemTemporaryDirectory_Darwin, u.settings.TempDir)

	case "linux": // Linux zoo.
		err = os.Setenv(EnvVar_SystemTemporaryDirectory_Linux, u.settings.TempDir)

	case "freebsd": // FreeBSD.
		err = os.Setenv(EnvVar_SystemTemporaryDirectory_FreeBSD, u.settings.TempDir)

	default: // Unknown O.S.
		return fmt.Errorf(Errf_UnknownOsType, runtime.GOOS)
	}

	if err != nil {
		return err
	}

	log.Println(fmt.Sprintf(Msgf_TemporaryDirectory, u.settings.TempDir))

	return nil
}
func (u *Uploader) Start() (err error) {
	u.controls.startStopMutex.Lock()
	defer u.controls.startStopMutex.Unlock()
	if u.controls.isRunning.Load() {
		return errors.New(Err_AlreadyStarted)
	}

	// 1. Start Watcher.
	err = u.startWatcher()
	if err != nil {
		return err
	}

	// 2.  Start HTTP Server.
	err = u.startHttpServer()
	if err != nil {
		return err
	}

	u.controls.isRunning.Store(true)
	return nil
}
func (u *Uploader) Stop() (err error) {
	u.controls.startStopMutex.Lock()
	defer u.controls.startStopMutex.Unlock()
	if !u.controls.isRunning.Load() {
		return errors.New(Err_AlreadyStopped)
	}

	log.Println(fmt.Sprintf(Msgf_FilesInQueue, u.controls.currentUploadsNum.Load()))

	// 1.  Stop HTTP Server.
	err = u.stopHttpServer()
	if err != nil {
		return err
	}

	// 2. Stop Watcher.
	err = u.stopWatcher()
	if err != nil {
		return err
	}

	u.controls.isRunning.Store(false)
	return nil
}
func (u *Uploader) IsRunning() bool {
	return u.controls.isRunning.Load()
}

func (u *Uploader) processFile(fileHeader *multipart.FileHeader, folderPath string, fileHash helper.HashSum, userName string, rw http.ResponseWriter) (err error) {
	if fileHeader == nil {
		u.httpRespond_BadRequest(rw)
		return errors.New(Err_FileHeaderIsNotAvailable)
	}

	var file multipart.File
	file, err = fileHeader.Open()
	if err != nil {
		u.httpRespond_InternalServerError(rw)
		return helper.CompositeError(Err_FileOperationError, err)
	}
	defer func() {
		derr := file.Close()
		if derr != nil {
			err = ers.Combine(err, derr)
		}
	}()

	fileName := fileHeader.Filename
	fileSize := fileHeader.Size

	var targetFilePath string
	if len(folderPath) == 0 {
		targetFilePath = filepath.Join(u.dataFolder, fileName)
	} else {
		targetFilePath = filepath.Join(u.dataFolder, folderPath, fileName)
	}

	var fileExists bool
	fileExists, err = af.FileExists(targetFilePath)
	if err != nil {
		u.httpRespond_InternalServerError(rw)
		return helper.CompositeError(Err_FileOperationError, err)
	}
	if fileExists {
		u.httpRespond_Conflict(rw)
		return fmt.Errorf(Errf_FileAlreadyExists, fileName)
	}

	u.controls.fileOperationMutex.Lock()
	defer u.controls.fileOperationMutex.Unlock()

	err = u.saveFile(targetFilePath, file)
	if err != nil {
		u.httpRespond_InternalServerError(rw)
		return helper.CompositeError(Err_FileOperationError, err)
	}

	var hashesMatch bool
	hashesMatch, err = u.verifyFile(targetFilePath, fileHash)
	if err != nil {
		u.httpRespond_InternalServerError(rw)
		return err
	}
	if !hashesMatch {
		err = u.deleteFile(targetFilePath)
		if err != nil {
			u.httpRespond_InternalServerError(rw)
			return helper.CompositeError(Err_FileOperationError, err)
		}

		u.httpRespond_BadRequest(rw)
		return fmt.Errorf(Errf_HashSumMismatch, fileName)
	}

	fileHashHexString := hex.EncodeToString(fileHash)
	err = u.updateJournal(userName, folderPath, int(fileSize), fileName, fileHashHexString)
	if err != nil {
		u.httpRespond_InternalServerError(rw)
		return err
	}

	return nil
}
func (u *Uploader) saveFile(targetFilePath string, file multipart.File) (err error) {
	// Create the folder if it does not exist.
	dir := filepath.Dir(targetFilePath)
	err = os.MkdirAll(dir, NewFolderPermissions)
	if err != nil {
		return err
	}

	// Create the file.
	var localFile *os.File
	localFile, err = os.Create(targetFilePath)
	if err != nil {
		return err
	}
	defer func() {
		derr := localFile.Close()
		if derr != nil {
			err = ers.Combine(err, derr)
		}
	}()

	_, err = io.Copy(localFile, file)
	if err != nil {
		return err
	}

	return nil
}
func (u *Uploader) verifyFile(targetFilePath string, hashSumSha256 []byte) (hashesMatch bool, err error) {
	var file *os.File
	file, err = os.Open(targetFilePath)
	if err != nil {
		return false, err
	}
	defer func() {
		derr := file.Close()
		if derr != nil {
			err = ers.Combine(err, derr)
		}
	}()

	hasher := sha256.New()

	_, err = io.Copy(hasher, file)
	if err != nil {
		return false, err
	}

	savedFileHashSum := hasher.Sum(nil)
	if !bytes.Equal(savedFileHashSum, hashSumSha256) {
		return false, nil
	}

	return true, nil
}
func (u *Uploader) deleteFile(targetFilePath string) (err error) {
	err = os.Remove(targetFilePath)
	if err != nil {
		return err
	}

	return nil
}
