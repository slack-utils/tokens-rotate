package fs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/slack-utils/tokens-rotate/internal/shared"
)

type Storage struct {
	shared.Token

	name string
}

var l *log.Entry

func (s *Storage) StorageGetName() string {
	return s.name
}

func (s *Storage) Read() error {
	f := viper.GetString("fs.token_file")

	if _, err := os.Stat(f); err != nil {
		return nil
	}

	data, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	if len(data) < 1 {
		return err
	}

	if err = json.Unmarshal(data, &s.Token); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Save() error {
	f := viper.GetString("fs.token_file")

	data, err := json.Marshal(s.Token)
	if err != nil {
		return err
	}

	if err := os.WriteFile(f, data, 0600); err != nil {
		return err
	}

	return nil
}

func (s *Storage) TokenGetAccess() (string, error) {
	return s.AccessToken, nil
}

func (s *Storage) TokenGetExpirationTime() (int64, error) {
	return s.Exp, nil
}

func (s *Storage) TokenGetRefresh() (string, error) {
	return s.RefreshToken, nil
}

func (s *Storage) TokenSetAccess(token string) string {
	s.AccessToken = token

	return token
}

func (s *Storage) TokenSetExpirationTime(exp int64) int64 {
	s.Exp = exp

	return exp
}

func (s *Storage) TokenSetRefresh(token string) string {
	s.RefreshToken = token

	return token
}

func New() *Storage {
	viper.SetDefault("fs.token_file", fmt.Sprintf("%s/token.json", shared.PathConf()))

	s := &Storage{name: "fs"}

	l = log.WithField("storage", s.name)

	return s
}
