package awssecrets

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/slack-utils/tokens-rotate/internal/shared"
)

type Storage struct {
	shared.Token

	name string
}

var (
	client     *secretsmanager.Client
	l          *log.Entry
	notFound   *types.ResourceNotFoundException
	secretName string
)

func (s *Storage) StorageGetName() string {
	return s.name
}

func (s *Storage) Read() error {
	s.init()

	if res, err := client.GetSecretValue(
		context.Background(),
		&secretsmanager.GetSecretValueInput{
			SecretId: &secretName,
		},
	); err != nil {
		if errors.As(err, &notFound) {
			secretString := "{}"

			if _, err := client.CreateSecret(
				context.Background(),
				&secretsmanager.CreateSecretInput{
					Name:         &secretName,
					SecretString: &secretString,
				},
			); err != nil {
				return err
			}

			return nil
		}

		return err
	} else {
		if err := json.Unmarshal([]byte(*res.SecretString), &s.Token); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) Save() error {
	s.init()

	data, err := json.Marshal(s.Token)
	if err != nil {
		return err
	}

	secretString := string(data)

	if _, err := client.PutSecretValue(
		context.Background(),
		&secretsmanager.PutSecretValueInput{
			SecretId:     &secretName,
			SecretString: &secretString,
		},
	); err != nil {
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
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			l.WithField("err", err).Fatal("failed get config")
		}

		client = secretsmanager.NewFromConfig(cfg)

		secretName = viper.GetString("awssecrets.secret_name")
	}
}

func New() *Storage {
	viper.SetDefault("awssecrets.secret_name", shared.PkgName)

	l = log.WithField("storage", "awssecrets")

	return &Storage{
		name: "awssecrets",
	}
}
