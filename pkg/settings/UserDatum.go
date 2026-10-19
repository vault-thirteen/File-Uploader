package settings

import "net/netip"

type UserDatum struct {
	Password  UserPassword
	IPAddress *netip.Addr
}
