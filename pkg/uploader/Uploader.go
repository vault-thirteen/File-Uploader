package uploader

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	cc "github.com/vault-thirteen/File-Uploader/pkg/CachedContent"
	"github.com/vault-thirteen/File-Uploader/pkg/helper"
	"github.com/vault-thirteen/File-Uploader/pkg/settings"

	ers "github.com/vault-thirteen/auxie/errors"
	af "github.com/vault-thirteen/auxie/file"
	"github.com/vault-thirteen/auxie/header"
	hh "github.com/vault-thirteen/auxie/http-helper"
	cci "github.com/vault-thirteen/auxie/http-helper/CachedContentItem"
)

const (
	NewFolderPermissions = 0777
	HttpContentCacheTime = 60
	URL_Upload           = "upload"
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
)

type Uploader struct {
	settings           *settings.Settings
	dataFolder         string
	httpServer         *http.Server
	cachedContent      *cc.CachedContent
	fileOperationMutex *sync.Mutex
}

func NewUploader(s *settings.Settings) (u *Uploader, err error) {
	u = &Uploader{
		settings:           s,
		dataFolder:         s.DataFolder,
		fileOperationMutex: &sync.Mutex{},
	}

	err = u.checkDataFolder()
	if err != nil {
		return nil, err
	}

	u.cachedContent, err = cc.NewCachedContent(s.AssetsFolder, HttpContentCacheTime)
	if err != nil {
		return nil, err
	}

	u.httpServer = &http.Server{
		Addr: net.JoinHostPort(
			s.Host,
			strconv.FormatUint(uint64(s.Port), 10),
		),
		Handler: http.HandlerFunc(u.router),
	}

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

func (u *Uploader) Run() (err error) {
	err = u.httpServer.ListenAndServeTLS(u.settings.SslCertFile, u.settings.SslKeyFile)
	if err != nil {
		return err
	}

	return nil
}

func (u *Uploader) router(rw http.ResponseWriter, req *http.Request) {
	left, right, ok := strings.Cut(req.URL.Path, helper.UrlPathSeparator)
	if !ok {
		u.httpRespond_BadRequest(rw)
		return
	}

	if len(left) != 0 {
		u.httpRespond_NotFound(rw)
		return
	}

	switch right {
	case "", cc.Asset_IndexHtml:
		u.httpRespond_CachedContent(rw, u.cachedContent.IndexHtml)
		return

	case cc.Asset_ScriptsJs:
		u.httpRespond_CachedContent(rw, u.cachedContent.ScriptsJs)
		return

	case cc.Asset_Sha256MinJs:
		u.httpRespond_CachedContent(rw, u.cachedContent.Sha256MinJs)
		return

	case cc.Asset_StylesCss:
		u.httpRespond_CachedContent(rw, u.cachedContent.StylesCss)
		return

	case cc.Asset_FaviconPng:
		u.httpRespond_CachedContent(rw, u.cachedContent.FaviconPng)
		return

	case URL_Upload:
		u.router_upload(rw, req)
		return

	default:
		u.httpRespond_NotFound(rw)
		return
	}
}
func (u *Uploader) router_upload(rw http.ResponseWriter, req *http.Request) {
	err := req.ParseMultipartForm(u.settings.FormSizeMax)
	if err != nil {
		u.httpRespond_BadRequest(rw)
		return
	}
	defer func() {
		derr := req.MultipartForm.RemoveAll()
		if derr != nil {
			err = ers.Combine(err, derr)
		}
	}()

	userName := req.FormValue(FormField_UserName)
	if len(userName) == 0 {
		u.httpRespond_BadRequest(rw)
		return
	}

	userPwd := req.FormValue(FormField_UserPassword)
	if len(userPwd) == 0 {
		u.httpRespond_BadRequest(rw)
		return
	}

	folderPath := req.FormValue(FormField_FolderPath)
	err = helper.CheckPath(folderPath)
	if err != nil {
		u.httpRespond_BadRequest(rw)
		return
	}

	var hashes []helper.HashSum
	hashes, err = helper.ParseFileHashes(req.FormValue(FormField_FileHashes))
	if err != nil {
		u.httpRespond_BadRequest(rw)
		return
	}

	var userIPAddressText string
	userIPAddressText, err = helper.GetClientIPAddress(req)
	if err != nil {
		u.httpRespond_BadRequest(rw)
		return
	}

	// Check the client.
	err = u.settings.CheckClient(userName, userPwd, userIPAddressText)
	if err != nil {
		u.httpRespond_Forbidden(rw)
		return
	}

	files := req.MultipartForm.File[FormField_Files]
	for i, fileHeader := range files {
		err = u.processFile(fileHeader, folderPath, hashes[i], userName, rw)
		if err != nil {
			log.Println(err)
			return
		}
	}

	return
}

func (u *Uploader) httpRespond_BadRequest(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusBadRequest)
}
func (u *Uploader) httpRespond_NotFound(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusNotFound)
}
func (u *Uploader) httpRespond_Forbidden(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusForbidden)
}
func (u *Uploader) httpRespond_Conflict(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusConflict)
}
func (u *Uploader) httpRespond_InternalServerError(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusInternalServerError)
}
func (u *Uploader) httpRespond_CachedContent(rw http.ResponseWriter, i *cci.CachedContentItem) {
	now := time.Now().UTC()
	rw.Header().Set(header.HttpHeaderContentType, i.ContentType())
	hh.SetCacheTime(rw, i.TTLSec(), now)

	_, err := rw.Write(i.Data())
	if err != nil {
		log.Println(err)
	}
}

func (u *Uploader) processFile(
	fileHeader *multipart.FileHeader,
	folderPath string,
	fileHash helper.HashSum,
	userName string,
	rw http.ResponseWriter,
) (err error) {
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

	msg := fmt.Sprintf("UserName: %s, FileSize: %d, FolderPath: \"%s\", FileName: \"%s\".", userName, fileSize, folderPath, fileName)
	log.Println(msg)

	u.fileOperationMutex.Lock()
	defer u.fileOperationMutex.Unlock()

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
