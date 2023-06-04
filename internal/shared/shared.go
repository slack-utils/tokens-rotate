package shared

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	Exp          int64  `json:"exp"`
	RefreshToken string `json:"refresh_token"`
}

var (
	PkgName = "tokens-rotate"
	Version = ""
)

func PathConf() string {
	ex, err := os.Executable()
	if err != nil {
		log.WithField("err", err).Fatal("failed get path name")
	}

	ExecutableDir := filepath.Dir(ex)

	pathConf, _ := filepath.Abs(ExecutableDir + "/../../../etc/" + PkgName)

	return pathConf
}
