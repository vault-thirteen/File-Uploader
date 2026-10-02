package uploader

import (
	"os"
	"strconv"
	"time"

	ers "github.com/vault-thirteen/auxie/errors"
)

const TimeFormat = "2006-01-02 15:04:05"

func (u *Uploader) updateJournal(userName string, folderPath string, fileSize int, fileName string, hash string) (err error) {
	var f *os.File
	f, err = os.OpenFile(u.settings.Journal, os.O_APPEND|os.O_CREATE|os.O_WRONLY, JournalFilePermissions)
	if err != nil {
		return err
	}
	defer func() {
		derr := f.Close()
		if derr != nil {
			err = ers.Combine(err, derr)
		}
	}()

	msg := formatJournalLine(userName, fileSize, folderPath, fileName, hash)

	_, err = f.WriteString(msg)
	if err != nil {
		return err
	}

	return nil
}
func formatJournalLine(userName string, fileSize int, folderPath string, fileName string, hash string) string {
	timeNow := time.Now().UTC()

	return timeNow.Format(TimeFormat) + JournalTabulator +
		userName + JournalTabulator +
		strconv.Itoa(fileSize) + JournalTabulator +
		folderPath + JournalTabulator +
		fileName + JournalTabulator +
		hash + JournalLineEnd
}
