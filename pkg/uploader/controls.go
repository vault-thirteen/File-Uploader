package uploader

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vault-thirteen/File-Uploader/pkg/helper"
)

const (
	Err_ServiceError = "service error"
)

const (
	Msg_WatcherStart      = "Watcher start"
	Msg_WatcherShutdown   = "Watcher shutdown"
	Msg_WatcherHasStopped = "Watcher has stopped"
)

type UploaderControls struct {
	startStopMutex     *sync.Mutex
	fileOperationMutex *sync.Mutex
	isRunning          *atomic.Bool
	serviceMustStop    *atomic.Bool
	watcherMustStop    *atomic.Bool
	errors             *chan error
	watcherWG          *sync.WaitGroup
	currentUploadsNum  *atomic.Int32
}

func NewUploaderControls() (uc *UploaderControls) {
	uc = &UploaderControls{
		startStopMutex:     new(sync.Mutex),
		fileOperationMutex: new(sync.Mutex),
		isRunning:          new(atomic.Bool),
		serviceMustStop:    new(atomic.Bool),
		watcherMustStop:    new(atomic.Bool),
		watcherWG:          new(sync.WaitGroup),
		currentUploadsNum:  new(atomic.Int32),
	}

	ec := make(chan error, 1)
	uc.errors = &ec

	uc.isRunning.Store(false)
	uc.serviceMustStop.Store(false)
	uc.watcherMustStop.Store(false)
	uc.currentUploadsNum.Store(0)

	return uc
}

func (u *Uploader) startWatcher() (err error) {
	log.Println(Msg_WatcherStart)

	u.controls.watcherWG.Add(1)
	go u.runWatcher()
	return nil
}
func (u *Uploader) stopWatcher() (err error) {
	log.Println(Msg_WatcherShutdown)

	u.controls.watcherMustStop.Store(true)
	u.controls.watcherWG.Wait()

	log.Println(Msg_WatcherHasStopped)

	return nil
}
func (u *Uploader) runWatcher() {
	defer u.controls.watcherWG.Done()

	var err error
	for {
		if u.controls.watcherMustStop.Load() {
			break
		}

		select {
		case err = <-*u.controls.errors:
			{
				log.Println(helper.CompositeError(Err_ServiceError, err))

				if errors.Is(err, http.ErrServerClosed) {
					break
				}

				go func() {
					err = u.Stop()
					if err != nil {
						log.Println(err)
					}
				}()
			}

		default:
			time.Sleep(time.Second)
		}
	}
}
