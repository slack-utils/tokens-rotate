package vault

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/slack-utils/tokens-rotate/internal/shared"
)

type Storage struct {
	shared.Token

	name string
}

var (
	client     *vault.Client
	l          *log.Entry
	secretName string
	secretPath string
)

func (s *Storage) StorageGetName() string {
	return s.name
}

func (s *Storage) Read() error {
	s.check()

	value, err := client.Secrets.KvV2Read(
		context.Background(),
		secretPath,
		vault.WithMountPath(secretName),
	)
	if err != nil {
		if vault.IsErrorStatus(err, http.StatusNotFound) {
			return nil
		}

		return err
	}

	s.Token.AccessToken = fmt.Sprintf("%s", value.Data.Data["access_token"])
	s.Token.RefreshToken = fmt.Sprintf("%s", value.Data.Data["refresh_token"])

	exp, err := strconv.ParseInt(fmt.Sprintf("%s", value.Data.Data["exp"]), 10, 64)
	if err != nil {
		return err
	}
	s.Token.Exp = exp

	return nil
}

func (s *Storage) Save() error {
	s.check()

	if _, err := client.Secrets.KvV2Write(
		context.Background(),
		secretPath,
		schema.KvV2WriteRequest{
			Data: map[string]any{
				"access_token":  s.Token.AccessToken,
				"exp":           fmt.Sprintf("%d", s.Token.Exp),
				"refresh_token": s.Token.RefreshToken,
			},
		},
		vault.WithMountPath(secretName),
	); err != nil {
		l.WithField("err", err).Fatal("failed to save secret")
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

func (s *Storage) TokenSetAccess(value string) string {
	s.AccessToken = value

	return value
}

func (s *Storage) TokenSetExpirationTime(value int64) int64 {
	s.Exp = value

	return value
}

func (s *Storage) TokenSetRefresh(value string) string {
	s.RefreshToken = value

	return value
}

func (s *Storage) init() {
	if client == nil {
		if c, err := vault.New(
			vault.WithEnvironment(),
		); err != nil {
			l.WithField("err", err).Fatal("failed get path name")
		} else {
			client = c
		}

		secretName = viper.GetString("vault.secret_name")
		secretPath = viper.GetString("vault.secret_path")
	}
}

func (s *Storage) check() {
	s.init()

	if _, err := client.System.ReadHealthStatus(
		context.Background(),
	); err != nil {
		l.WithField("err", err).Fatal("health check was failed")
	}

	if _, err := client.Auth.TokenLookUpSelf(
		context.Background(),
	); err != nil {
		l.WithField("err", err).Fatal("token lookup was failed")
	}
}

func New() *Storage {
	viper.SetDefault("vault.secret_name", "secret")
	viper.SetDefault("vault.secret_path", shared.PkgName)

	l = log.WithField("storage", "vault")

	return &Storage{
		name: "vault",
	}
}
