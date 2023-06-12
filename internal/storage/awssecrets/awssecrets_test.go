package awssecrets

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/stretchr/testify/mock"
)

type ClientMock struct {
	mock.Mock
}

func (c *ClientMock) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	args := c.Called(ctx, params, []func(*secretsmanager.Options){})
	return args.Get(0).(*secretsmanager.GetSecretValueOutput), args.Error(1)
}
func (c *ClientMock) CreateSecret(ctx context.Context, params *secretsmanager.CreateSecretInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.CreateSecretOutput, error) {
	args := c.Called(ctx, params, []func(*secretsmanager.Options){})
	return args.Get(0).(*secretsmanager.CreateSecretOutput), args.Error(1)
}
func (c *ClientMock) PutSecretValue(ctx context.Context, params *secretsmanager.PutSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.PutSecretValueOutput, error) {
	args := c.Called(ctx, params, []func(*secretsmanager.Options){})
	return args.Get(0).(*secretsmanager.PutSecretValueOutput), args.Error(1)
}

func TestStorage(t *testing.T) {
	tests := []struct {
		name         string
		secretName   string
		secretString string
		prepare      func(*ClientMock, string, string)
	}{
		{
			name:         "reading the secret",
			secretName:   "test-secret-name",
			secretString: `{"access_token":"access-token","exp":0,"refresh_token":"refresh-token"}`,
			prepare: func(c *ClientMock, secretName, secretString string) {
				c.On(
					"GetSecretValue",
					context.Background(),
					&secretsmanager.GetSecretValueInput{SecretId: &secretName},
					[]func(*secretsmanager.Options){},
				).Return(&secretsmanager.GetSecretValueOutput{SecretString: &secretString}, nil)
			},
		},
		{
			name:         "creating the secret",
			secretName:   "test-secret-name",
			secretString: `{"access_token":"","exp":0,"refresh_token":""}`,
			prepare: func(c *ClientMock, secretName, secretString string) {
				emptySecret := "{}"

				c.On(
					"GetSecretValue",
					context.Background(),
					&secretsmanager.GetSecretValueInput{SecretId: &secretName},
					[]func(*secretsmanager.Options){},
				).Return(&secretsmanager.GetSecretValueOutput{}, &types.ResourceNotFoundException{})

				c.On(
					"CreateSecret",
					context.Background(),
					&secretsmanager.CreateSecretInput{Name: &secretName, SecretString: &emptySecret},
					[]func(*secretsmanager.Options){},
				).Return(&secretsmanager.CreateSecretOutput{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ClientMock{}

			s := &Storage{
				name:       "test",
				client:     c,
				secretName: tt.secretName,
			}

			if tt.prepare != nil {
				tt.prepare(c, tt.secretName, tt.secretString)
			}

			c.On(
				"PutSecretValue",
				context.Background(),
				&secretsmanager.PutSecretValueInput{SecretId: &tt.secretName, SecretString: &tt.secretString},
				[]func(*secretsmanager.Options){},
			).Return(&secretsmanager.PutSecretValueOutput{}, nil)

			s.StorageGetName()
			s.Read()
			s.Save()
			c.AssertExpectations(t)
		})
	}
}
