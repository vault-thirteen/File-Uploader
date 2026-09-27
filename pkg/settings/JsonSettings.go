package settings

// JsonSettings is a type for settings' raw data in JSON file.
type JsonSettings struct {
	Host         string             `json:"host"`
	Port         uint16             `json:"port"`
	SslCertFile  string             `json:"certFile"`
	SslKeyFile   string             `json:"keyFile"`
	DataFolder   string             `json:"dataFolder"`
	AssetsFolder string             `json:"assetsFolder"`
	Users        []JsonSettingsUser `json:"users"`
}

type JsonSettingsUser struct {
	Name      string `json:"name"`
	Password  string `json:"password"`
	IPAddress string `json:"ipAddress"`
}
