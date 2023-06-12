package awssecrets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/slack-utils/tokens-rotate/internal/shared"
)

type Client interface {
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
	CreateSecret(context.Context, *secretsmanager.CreateSecretInput, ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error)
	PutSecretValue(context.Context, *secretsmanager.PutSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error)
}

type Storage struct {
	shared.GeneralStorage

	client     Client
	l          *log.Entry
	name       string
	secretName string
}

var notFound *types.ResourceNotFoundException

func (s *Storage) StorageGetName() string {
	return s.name
}

func (s *Storage) Read() error {
	if res, err := s.client.GetSecretValue(
		context.Background(),
		&secretsmanager.GetSecretValueInput{
			SecretId: &s.secretName,
		},
	); err != nil {
		if errors.As(err, &notFound) {
			secretString := "{}"

			if _, err := s.client.CreateSecret(
				context.Background(),
				&secretsmanager.CreateSecretInput{
					Name:         &s.secretName,
					SecretString: &secretString,
				},
			); err != nil {
				return err
			}

			return fmt.Errorf("secret not fount")
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
	data, err := json.Marshal(s.Token)
	if err != nil {
		return err
	}

	secretString := string(data)

	if _, err := s.client.PutSecretValue(
		context.Background(),
		&secretsmanager.PutSecretValueInput{
			SecretId:     &s.secretName,
			SecretString: &secretString,
		},
	); err != nil {
		return err
	}

	return nil
}

func NewClient() Client {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.WithField("err", err).Fatal("failed get config")
	}

	return secretsmanager.NewFromConfig(cfg)
}

func New() *Storage {
	viper.SetDefault("awssecrets.secret_name", shared.PkgName)

	s := &Storage{
		client:     NewClient(),
		l:          log.WithField("storage", "awssecrets"),
		name:       "awssecrets",
		secretName: viper.GetString("awssecrets.secret_name"),
	}

	return s
}
