package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"os"

	ers "github.com/vault-thirteen/auxie/errors"
	af "github.com/vault-thirteen/auxie/file"
)

const (
	Err_PortIsNotSet             = "port is not set"
	Err_UsersAreNotSet           = "users are not set"
	Err_FormSizeIsNotSet         = "form size is not set"
	Err_DataFolderIsNotSet       = "folder is not set"
	Err_DataFolderDoesNotExist   = "folder does not exist"
	Err_AssetsFolderIsNotSet     = "assets folder is not set"
	Err_UserNameIsNotSet         = "user name is not set"
	Err_UserPasswordIsNotSet     = "user password is not set"
	Err_UserIPAddressIsNotSet    = "user IP address is not set"
	Errf_UserIPAddressIsNotValid = " IP address is not valid, user name: %s"
	Err_UserDataIsNotInitialised = "user data is not initialised"
)

const (
	UserIPAddressAny = `*`
)

type Settings struct {
	Host         string
	Port         uint16
	SslCertFile  string
	SslKeyFile   string
	FormSizeMax  int64
	DataFolder   string
	AssetsFolder string
	UserData     *UserData
}

func NewSettingsFromFile(settingsFilePath string) (s *Settings, err error) {
	var file *os.File
	file, err = os.Open(settingsFilePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		derr := file.Close()
		if derr == nil {
			err = ers.Combine(err, derr)
		}
	}()

	var rawData JsonSettings
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&rawData)
	if err != nil {
		return nil, err
	}

	s, err = NewSettingsFromRawData(rawData)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func NewSettingsFromRawData(rawData JsonSettings) (s *Settings, err error) {
	if rawData.Port == 0 {
		return nil, errors.New(Err_PortIsNotSet)
	}

	if len(rawData.Users) < 1 {
		return nil, errors.New(Err_UsersAreNotSet)
	}

	if rawData.FormSizeMax == 0 {
		return nil, errors.New(Err_FormSizeIsNotSet)
	}

	if len(rawData.DataFolder) < 1 {
		return nil, errors.New(Err_DataFolderIsNotSet)
	}

	var folderExists bool
	folderExists, err = af.FolderExists(rawData.DataFolder)
	if err != nil {
		return nil, err
	}
	if !folderExists {
		return nil, errors.New(Err_DataFolderDoesNotExist)
	}

	if len(rawData.AssetsFolder) < 1 {
		return nil, errors.New(Err_AssetsFolderIsNotSet)
	}

	ud := NewUserData()

	for _, user := range rawData.Users {
		if len(user.Name) == 0 {
			return nil, errors.New(Err_UserNameIsNotSet)
		}

		if len(user.Password) == 0 {
			return nil, errors.New(Err_UserPasswordIsNotSet)
		}

		if len(user.IPAddress) == 0 {
			return nil, errors.New(Err_UserIPAddressIsNotSet)
		}

		if user.IPAddress != UserIPAddressAny {
			_, err = netip.ParseAddr(user.IPAddress)
			if err != nil {
				return nil, fmt.Errorf(Errf_UserIPAddressIsNotValid, user.Name)
			}
		}

		err = ud.AddUser(user.Name, user.Password, user.IPAddress)
		if err != nil {
			return nil, err
		}

	}

	s = &Settings{
		Host:         rawData.Host,
		Port:         rawData.Port,
		SslCertFile:  rawData.SslCertFile,
		SslKeyFile:   rawData.SslKeyFile,
		FormSizeMax:  rawData.FormSizeMax,
		DataFolder:   rawData.DataFolder,
		AssetsFolder: rawData.AssetsFolder,
		UserData:     ud,
	}

	return s, nil
}

func (s *Settings) CheckUser(userName string, userPassword string, userIPAddress string) (err error) {
	if s.UserData == nil {
		return errors.New(Err_UserDataIsNotInitialised)
	}

	err = s.UserData.CheckUser(userName, userPassword, userIPAddress)
	if err != nil {
		return err
	}

	return nil
}
