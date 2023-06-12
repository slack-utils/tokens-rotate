package vault

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/slack-utils/tokens-rotate/internal/shared"
)

type Auth interface {
	TokenLookUpSelf(context.Context, ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
}

type System interface {
	ReadHealthStatus(context.Context, ...vault.RequestOption) (*vault.Response[map[string]interface{}], error)
}

type Secrets interface {
	KvV2Read(context.Context, string, ...vault.RequestOption) (*vault.Response[schema.KvV2ReadResponse], error)
	KvV2Write(context.Context, string, schema.KvV2WriteRequest, ...vault.RequestOption) (*vault.Response[schema.KvV2WriteResponse], error)
}

type Storage struct {
	shared.GeneralStorage

	auth    Auth
	secrets Secrets
	system  System

	l          *log.Entry
	name       string
	secretName string
	secretPath string
}

func (s *Storage) StorageGetName() string {
	return s.name
}

func (s *Storage) Read() error {
	s.check()

	value, err := s.secrets.KvV2Read(
		context.Background(),
		s.secretPath,
		vault.WithMountPath(s.secretName),
	)
	if err != nil {
		return err
	}

	s.Token.AccessToken = value.Data.Data["access_token"].(string)
	s.Token.RefreshToken = value.Data.Data["refresh_token"].(string)

	exp, err := strconv.ParseInt(value.Data.Data["exp"].(string), 10, 64)
	if err != nil {
		return err
	}
	s.Token.Exp = exp

	return nil
}

func (s *Storage) Save() error {
	s.check()

	if _, err := s.secrets.KvV2Write(
		context.Background(),
		s.secretPath,
		schema.KvV2WriteRequest{
			Data: map[string]any{
				"access_token":  s.Token.AccessToken,
				"exp":           fmt.Sprintf("%d", s.Token.Exp),
				"refresh_token": s.Token.RefreshToken,
			},
		},
		vault.WithMountPath(s.secretName),
	); err != nil {
		s.l.WithField("err", err).Fatal("failed to save secret")
	}

	return nil
}

func (s *Storage) check() {
	if _, err := s.system.ReadHealthStatus(
		context.Background(),
	); err != nil {
		s.l.WithField("err", err).Fatal("health check was failed")
	}

	if _, err := s.auth.TokenLookUpSelf(
		context.Background(),
	); err != nil {
		s.l.WithField("err", err).Fatal("token lookup was failed")
	}
}

func NewClient() *vault.Client {
	if c, err := vault.New(
		vault.WithEnvironment(),
	); err != nil {
		log.WithField("err", err).Fatal("failed get path name")
	} else {
		return c
	}

	return nil
}

func New() *Storage {
	viper.SetDefault("vault.secret_name", "secret")
	viper.SetDefault("vault.secret_path", shared.PkgName)

	c := NewClient()

	s := &Storage{
		auth:    &c.Auth,
		system:  &c.System,
		secrets: &c.Secrets,

		l:          log.WithField("storage", "vault"),
		name:       "vault",
		secretName: viper.GetString("vault.secret_name"),
		secretPath: viper.GetString("vault.secret_path"),
	}

	return s
}
