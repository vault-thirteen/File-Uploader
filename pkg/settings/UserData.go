package settings

import (
	"net/netip"

	"github.com/vault-thirteen/auxie/errors"
)

type UserData struct {
	data map[UserName]*UserDatum
}

const (
	Err_UserIsNotFound = "user is not found"
	Err_Authentication = "authentication error"

	Errf_UserAlreadyExists = "user already exists: %s"
)

func NewUserData() (ud *UserData) {
	return &UserData{
		data: make(map[UserName]*UserDatum),
	}
}

func (ud *UserData) AddUser(userName string, userPassword string, userIPAddress string) (err error) {
	newUser := &UserDatum{
		Password: UserPassword{
			text: &userPassword,
		},
	}

	if userIPAddress == UserIPAddressAny {
		newUser.IPAddress = nil
	} else {
		var ipa netip.Addr
		ipa, err = netip.ParseAddr(userIPAddress)
		if err != nil {
			return err
		}

		newUser.IPAddress = &ipa
	}

	// Check for duplicates.
	var exists bool
	_, exists = ud.data[UserName(userName)]
	if exists {
		return errors.ErrorFromFormattedText(Errf_UserAlreadyExists, userName)
	}

	ud.data[UserName(userName)] = newUser

	return nil
}

func (ud *UserData) CheckClient(userName string, userPassword string, userIPAddress string) (err error) {
	userData, exists := ud.data[UserName(userName)]
	if !exists {
		return errors.ErrorFromText(Err_UserIsNotFound)
	}
	if userData == nil {
		return errors.ErrorFromText(Err_UserIsNotFound)
	}

	if *userData.Password.text != userPassword {
		return errors.ErrorFromText(Err_Authentication)
	}

	if userData.IPAddress != nil {
		var ipa netip.Addr
		ipa, err = netip.ParseAddr(userIPAddress)
		if err != nil {
			return err
		}

		if *userData.IPAddress != ipa {
			return errors.ErrorFromText(Err_Authentication)
		}
	}

	return nil
}
