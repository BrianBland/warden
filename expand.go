package warden

import (
	"os"
)

func expand(path string) string {
	if path[:2] == "~/" {
		path = "$HOME" + path[1:]
	}
	return os.ExpandEnv(path)
}
