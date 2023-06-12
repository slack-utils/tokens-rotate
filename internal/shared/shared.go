package shared

import (
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	Exp          int64  `json:"exp"`
	RefreshToken string `json:"refresh_token"`
}

type GeneralStorage struct {
	Token
}

func (gs *GeneralStorage) TokenGetAccess() string {
	return gs.AccessToken
}

func (gs *GeneralStorage) TokenGetExpirationTime() int64 {
	return gs.Exp
}

func (gs *GeneralStorage) TokenGetRefresh() string {
	return gs.RefreshToken
}

func (gs *GeneralStorage) TokenSetAccess(token string) {
	gs.AccessToken = token
}

func (gs *GeneralStorage) TokenSetExpirationTime(exp int64) {
	gs.Exp = exp
}

func (gs *GeneralStorage) TokenSetRefresh(token string) {
	gs.RefreshToken = token
}

func (gs *GeneralStorage) LoadTokensFromEnv() {
	log.Info("importing tokens from environment variables")

	gs.TokenSetAccess(viper.GetString("access_token"))
	gs.TokenSetRefresh(viper.GetString("refresh_token"))
	gs.TokenSetExpirationTime(
		time.
			Now().
			Add(time.Hour * time.Duration(12)).
			Unix(),
	)
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
