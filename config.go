package warden

type Config struct {
	Addr        string   `json:"addr"`
	PrivateKeys []string `json:"privateKeys"`
	Jail        Jail     `json:"jail"`
}

type Jail struct {
	Image      string `json:"image"`
	Persistent bool   `json:"persistent"`
}
