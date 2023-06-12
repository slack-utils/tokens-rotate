package vault

import (
	"context"
	"testing"

	"github.com/hashicorp/vault-client-go"
	"github.com/hashicorp/vault-client-go/schema"
	"github.com/stretchr/testify/mock"
)

type AuthMock struct {
	mock.Mock
}

func (a *AuthMock) TokenLookUpSelf(ctx context.Context, _ ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	args := a.Called(ctx)
	return args.Get(0).(*vault.Response[map[string]interface{}]), args.Error(1)
}

type SystemMock struct {
	mock.Mock
}

func (s *SystemMock) ReadHealthStatus(ctx context.Context, _ ...vault.RequestOption) (*vault.Response[map[string]interface{}], error) {
	args := s.Called(ctx)
	return args.Get(0).(*vault.Response[map[string]interface{}]), args.Error(1)
}

type SecretsMock struct {
	mock.Mock
}

func (s *SecretsMock) KvV2Read(ctx context.Context, path string, _ ...vault.RequestOption) (*vault.Response[schema.KvV2ReadResponse], error) {
	args := s.Called(
		ctx,
		path,
	)
	return args.Get(0).(*vault.Response[schema.KvV2ReadResponse]), args.Error(1)
}
func (s *SecretsMock) KvV2Write(ctx context.Context, path string, request schema.KvV2WriteRequest, _ ...vault.RequestOption) (*vault.Response[schema.KvV2WriteResponse], error) {
	args := s.Called(ctx, path, request)
	return args.Get(0).(*vault.Response[schema.KvV2WriteResponse]), args.Error(1)
}

func TestStorage(t *testing.T) {
	secret_name := "secret-name"
	secret_path := "secret-path"
	token := map[string]interface{}{
		"access_token":  "access-token",
		"refresh_token": "refresh-token",
		"exp":           "123",
	}

	auth := &AuthMock{}
	system := &SystemMock{}
	secrets := &SecretsMock{}

	auth.On(
		"TokenLookUpSelf",
		context.Background(),
	).Return(&vault.Response[map[string]interface{}]{}, nil)

	system.On(
		"ReadHealthStatus",
		context.Background(),
	).Return(&vault.Response[map[string]interface{}]{}, nil)

	secrets.On(
		"KvV2Read",
		context.Background(),
		secret_path,
	).Return(&vault.Response[schema.KvV2ReadResponse]{
		Data: schema.KvV2ReadResponse{Data: token},
	}, nil)

	secrets.On(
		"KvV2Write",
		context.Background(),
		secret_path,
		schema.KvV2WriteRequest{Data: token},
	).Return(&vault.Response[schema.KvV2WriteResponse]{}, nil)

	s := &Storage{
		auth:    auth,
		system:  system,
		secrets: secrets,

		name:       "test",
		secretName: secret_name,
		secretPath: secret_path,
	}

	s.StorageGetName()
	s.Read()
	s.Save()

	auth.AssertExpectations(t)
	system.AssertExpectations(t)
	secrets.AssertExpectations(t)
}
