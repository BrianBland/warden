package warden

type Config struct {
	Addr        string   `json:"addr"`
	PrivateKeys []string `json:"privateKeys"`
	JailImage   string   `json:"jailImage"`
}
