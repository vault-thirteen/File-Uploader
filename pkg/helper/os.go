package helper

import (
	"strings"

	"github.com/vault-thirteen/auxie/errors"
)

const (
	Err_PathIsNotValid = "path is not valid"
)

func CheckPath(path string) (err error) {
	if strings.Contains(path, "..") {
		return errors.ErrorFromText(Err_PathIsNotValid)
	}

	return nil
}
