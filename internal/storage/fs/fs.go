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
	shared.GeneralStorage

	l          *log.Entry
	name       string
	token_file string
}

func (s *Storage) StorageGetName() string {
	return s.name
}

func (s *Storage) Read() error {
	if _, err := os.Stat(s.token_file); err != nil {
		return fmt.Errorf("token file not fount")
	}

	data, err := ioutil.ReadFile(s.token_file)
	if err != nil {
		return fmt.Errorf("cant read token file")
	}

	if len(data) < 1 {
		return err
	}

	if err = json.Unmarshal(data, &s.Token); err != nil {
		return err
	}

	s.l.Debugf("tokens - %#v", s.Token)

	return nil
}

func (s *Storage) Save() error {
	data, err := json.Marshal(s.Token)
	if err != nil {
		return err
	}

	if err := os.WriteFile(s.token_file, data, 0600); err != nil {
		return err
	}

	return nil
}

func New() *Storage {
	viper.SetDefault("fs.token_file", fmt.Sprintf("%s/token.json", shared.PathConf()))

	s := &Storage{
		l:          log.WithField("storage", "fs"),
		name:       "fs",
		token_file: viper.GetString("fs.token_file"),
	}

	return s
}
