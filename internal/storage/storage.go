package storage

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/spf13/viper"

	"github.com/slack-utils/tokens-rotate/internal/storage/awssecrets"
	"github.com/slack-utils/tokens-rotate/internal/storage/fs"
	"github.com/slack-utils/tokens-rotate/internal/storage/vault"
)

type Storage interface {
	Read() error
	Save() error
	StorageGetName() string
	TokenGetAccess() (string, error)
	TokenGetExpirationTime() (int64, error)
	TokenGetRefresh() (string, error)
	TokenSetAccess(string) string
	TokenSetExpirationTime(int64) int64
	TokenSetRefresh(string) string
}

var (
	api         *slack.Client
	err         error
	storages    = map[string]Storage{}
	retry_limit = 3

	access_token    string
	expiration_time int64
	refresh_token   string
	storage         string
)

func Run() {
	go rotation(pre())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigs

	log.WithFields(log.Fields{
		"signal": sig.String(),
		"code":   fmt.Sprintf("%d", sig),
	}).Info("Signal notify")
}

func pre() Storage {
	if storage = viper.GetString("storage"); storages[storage] == nil {
		log.WithField("storage", storage).Fatal("Unknown storage")
	}

	s := storages[storage]

	if err := s.Read(); err != nil {
		log.WithField("err", err).Fatal("failed to read token from storage")
	}

	if access_token, err = s.TokenGetAccess(); err != nil {
		log.Fatalf("failed to get access token: %s", err)
	} else if access_token == "" {
		access_token = viper.GetString("access_token")
	}

	if refresh_token, err = s.TokenGetRefresh(); err != nil {
		log.Fatalf("failed to get refresh token: %s", err)
	} else if refresh_token == "" {
		refresh_token = viper.GetString("refresh_token")
	}

	if expiration_time, err = s.TokenGetExpirationTime(); err != nil {
		log.Fatalf("failed to get expiration time: %s", err)
	} else if expiration_time == 0 {
		expiration_time = time.Now().Add(time.Hour * time.Duration(12)).Unix()
	}

	if access_token == "" || refresh_token == "" {
		log.Fatal("access or refresh token is empty")
	}

	return s
}

func rotation(s Storage) {
	api = slack.New(access_token)

	log.Info("checking the current token")
	if _, err := api.AuthTest(); err != nil {
		log.Errorf("the current token is invalid: %s", err)
		access_token = ""
	}

	for {
		if access_token != "" {
			log.Infof("next rotation in %s", time.Unix(expiration_time, 0))
			time.Sleep(time.Second * time.Duration(expiration_time-time.Now().Add(time.Hour).Unix()))
		}

		log.Info("rotation of the token")
		token, err := api.ToolingTokensRotate(refresh_token)
		if err != nil {
			log.WithField("err", err).Error("failed to rotate token")

			refresh_token = viper.GetString("refresh_token")

			retry_limit -= 1

			if retry_limit < 1 {
				os.Exit(1)
			}

			time.Sleep(time.Second)

			continue
		}

		access_token = s.TokenSetAccess(token.Token)
		expiration_time = s.TokenSetExpirationTime(token.Exp)
		refresh_token = s.TokenSetRefresh(token.RefreshToken)

		if err := s.Save(); err != nil {
			log.WithField("err", err).Error()

			continue
		}

		api = slack.New(access_token)
	}
}

func init() {
	list := []Storage{
		awssecrets.New(),
		fs.New(),
		vault.New(),
	}

	for _, item := range list {
		name := item.StorageGetName()

		storages[name] = item
	}
}
