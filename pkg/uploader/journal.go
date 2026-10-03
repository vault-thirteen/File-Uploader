package uploader

import (
	"strconv"
	"time"
)

const TimeFormat = "2006-01-02 15:04:05"

func (u *Uploader) updateJournal(userName string, folderPath string, fileSize int, fileName string, hash string) (err error) {
	msg := formatJournalLine(userName, fileSize, folderPath, fileName, hash)

	_, err = u.journalFile.WriteString(msg)
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
