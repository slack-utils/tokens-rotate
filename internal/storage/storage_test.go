//go:generate mockgen -source=${GOFILE} -destination=mock/${GOFILE}
package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/mock"
)

type StorageMock struct {
	mock.Mock
}

func (s *StorageMock) LoadTokensFromEnv() {
	s.Called()
}
func (s *StorageMock) Read() error {
	args := s.Called()
	return args.Error(0)
}
func (s *StorageMock) Save() error {
	args := s.Called()
	return args.Error(0)
}
func (s *StorageMock) StorageGetName() string {
	args := s.Called()
	return args.Get(0).(string)
}
func (s *StorageMock) TokenGetAccess() string {
	args := s.Called()
	return args.Get(0).(string)
}
func (s *StorageMock) TokenGetExpirationTime() int64 {
	args := s.Called()
	return args.Get(0).(int64)
}
func (s *StorageMock) TokenGetRefresh() string {
	args := s.Called()
	return args.Get(0).(string)
}
func (s *StorageMock) TokenSetAccess(token string) {
	s.Called(token)
}
func (s *StorageMock) TokenSetExpirationTime(exp int64) {
	s.Called(exp)
}
func (s *StorageMock) TokenSetRefresh(token string) {
	s.Called(token)
}

type SlackMock struct {
	mock.Mock
}

func (s *SlackMock) AuthTest() (*slack.AuthTestResponse, error) {
	args := s.Called()
	return args.Get(0).(*slack.AuthTestResponse), args.Error(1)
}
func (s *SlackMock) ToolingTokensRotate(token string) (*slack.ToolingTokensRotate, error) {
	args := s.Called(token)
	return args.Get(0).(*slack.ToolingTokensRotate), args.Error(1)
}

func TestNewSlack(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*StorageMock, *SlackMock, *slack.ToolingTokensRotate)
	}{
		{
			name: "working token",
			prepare: func(s *StorageMock, c *SlackMock, token *slack.ToolingTokensRotate) {
				s.On("TokenGetAccess").Return("test-access-token")
				c.On("AuthTest").Return(&slack.AuthTestResponse{}, nil)
			},
		},
		{
			name: "renew token",
			prepare: func(s *StorageMock, c *SlackMock, token *slack.ToolingTokensRotate) {
				s.On("TokenGetAccess").Return("test-access-token")
				s.On("TokenGetRefresh").Return("test-refresh-token")
				s.On("LoadTokensFromEnv").Return()
				s.On("TokenSetAccess", token.Token).Return()
				s.On("TokenSetRefresh", token.RefreshToken).Return()
				s.On("TokenSetExpirationTime", token.Exp).Return()
				s.On("Save").Return(nil)

				c.On("AuthTest").Return(&slack.AuthTestResponse{}, fmt.Errorf("invalid_auth"))
				c.On("ToolingTokensRotate", "test-refresh-token").Return(&slack.ToolingTokensRotate{}, fmt.Errorf("invalid_refresh_token")).Times(2)
				c.On("ToolingTokensRotate", "test-refresh-token").Return(token, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StorageMock{}
			c := &SlackMock{}

			if tt.prepare != nil {
				tt.prepare(s, c, &slack.ToolingTokensRotate{
					Exp:          123,
					RefreshToken: "new-refresh-token",
					Token:        "new-access-token",
				})
			}

			a := NewSlack(s, func(_ string, _ ...slack.Option) SlackClient {
				return c
			})

			ctx, stop := context.WithTimeout(
				context.Background(),
				time.Millisecond*1000,
			)
			defer stop()

			a.Run(ctx, time.Second)
			s.AssertExpectations(t)
		})
	}
}
