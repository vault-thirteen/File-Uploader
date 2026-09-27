package helper

import (
	"errors"
	"strings"
)

const (
	Err_PathIsNotValid = "path is not valid"
)

func CheckPath(path string) (err error) {
	if strings.Contains(path, "..") {
		return errors.New(Err_PathIsNotValid)
	}

	return nil
}
