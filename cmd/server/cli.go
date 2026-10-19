package main

import (
	"os"

	"github.com/vault-thirteen/auxie/errors"
)

const OS_Args_Count = 1

const (
	Err_ArgumentsCount       = "not enough arguments"
	Err_SettingsFileIsNotSet = "settings file is not set"
)

type Arguments struct {
	SettingsFilePath string
}

func (a *Arguments) IsValid() (ok bool, err error) {
	if len(a.SettingsFilePath) < 1 {
		return false, errors.ErrorFromText(Err_SettingsFileIsNotSet)
	}

	return true, nil
}

func NewArgumentsFromOs() (args *Arguments, err error) {
	if len(os.Args) != OS_Args_Count+1 {
		return nil, errors.ErrorFromText(Err_ArgumentsCount)
	}

	args = &Arguments{
		SettingsFilePath: os.Args[1],
	}

	_, err = args.IsValid()
	if err != nil {
		return nil, err
	}

	return args, nil
}
