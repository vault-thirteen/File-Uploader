package uploader

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	mime "github.com/vault-thirteen/auxie/MIME"
	ae "github.com/vault-thirteen/auxie/errors"
	"github.com/vault-thirteen/auxie/header"
	hh "github.com/vault-thirteen/auxie/http-helper"
	cci "github.com/vault-thirteen/auxie/http-helper/CachedContentItem"

	cc "github.com/vault-thirteen/File-Uploader/pkg/CachedContent"
	"github.com/vault-thirteen/File-Uploader/pkg/helper"
)

const (
	Msg_HttpServerStart      = "HTTP server start"
	Msg_HttpServerShutdown   = "HTTP server shutdown"
	Msg_HttpServerHasStopped = "HTTP server has stopped"
)

func (u *Uploader) NewHttpServer(host string, port uint16) (hs *http.Server) {
	return &http.Server{
		Addr: net.JoinHostPort(
			host,
			strconv.FormatUint(uint64(port), 10),
		),
		Handler: http.HandlerFunc(u.router),
	}
}
func (u *Uploader) startHttpServer() (err error) {
	log.Println(Msg_HttpServerStart)

	go u.runHttpServer()
	return nil
}
func (u *Uploader) stopHttpServer() (err error) {
	log.Println(Msg_HttpServerShutdown)

	err = u.httpServer.Shutdown(context.Background())
	if err != nil {
		return err
	}

	log.Println(Msg_HttpServerHasStopped)

	return nil
}
func (u *Uploader) runHttpServer() {
	err := u.httpServer.ListenAndServeTLS(u.settings.SslCertFile, u.settings.SslKeyFile)
	if err != nil {
		*u.controls.errors <- err
	}
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
	case URL_QueueSize:
		u.router_queueSize(rw, req)
		return

	case URL_Upload:
		u.router_upload(rw, req)
		return

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

	default:
		u.httpRespond_NotFound(rw)
		return
	}
}
func (u *Uploader) router_queueSize(rw http.ResponseWriter, req *http.Request) {
	q := Queue{Size: int(u.controls.currentUploadsNum.Load())}
	u.httpRespond_JsonObject(rw, q)
	return
}
func (u *Uploader) router_upload(rw http.ResponseWriter, req *http.Request) {
	currentUploadsCount := int(u.controls.currentUploadsNum.Load())
	if currentUploadsCount >= u.settings.SimultaneousUploadsCount {
		u.httpRespond_ServiceUnavailable(rw)
		return
	}

	u.controls.currentUploadsNum.Add(1)
	defer u.controls.currentUploadsNum.Add(-1)

	err := req.ParseMultipartForm(u.settings.FormSizeMax)
	if err != nil {
		u.httpRespond_BadRequest(rw)
		return
	}
	defer func() {
		derr := req.MultipartForm.RemoveAll()
		if derr != nil {
			err = ae.Combine(err, derr)
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
func (u *Uploader) httpRespond_ServiceUnavailable(rw http.ResponseWriter) {
	rw.WriteHeader(http.StatusServiceUnavailable)
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
func (u *Uploader) httpRespond_JsonObject(rw http.ResponseWriter, obj any) {
	rw.Header().Set(header.HttpHeaderContentType, mime.TypeApplicationJson)

	err := json.NewEncoder(rw).Encode(obj)
	if err != nil {
		log.Println(err)
		return
	}
}
