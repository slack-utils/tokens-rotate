//go:generate mockgen -source=${GOFILE} -destination=mock/${GOFILE}
package storage

import (
	"context"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"

	"github.com/slack-utils/tokens-rotate/internal/storage/awssecrets"
	"github.com/slack-utils/tokens-rotate/internal/storage/fs"
	"github.com/slack-utils/tokens-rotate/internal/storage/vault"
)

type Storage interface {
	LoadTokensFromEnv()
	Read() error
	Save() error
	StorageGetName() string
	TokenGetAccess() string
	TokenGetExpirationTime() int64
	TokenGetRefresh() string
	TokenSetAccess(string)
	TokenSetExpirationTime(int64)
	TokenSetRefresh(string)
}

type SlackClient interface {
	AuthTest() (*slack.AuthTestResponse, error)
	ToolingTokensRotate(refresh_token string) (*slack.ToolingTokensRotate, error)
}

type SlackClientFactory func(string, ...slack.Option) SlackClient

type App struct {
	SlackClient
	Storage

	factory SlackClientFactory
}

func (a *App) tokenRotate() {
	log.Info("rotating token")
	token, err := a.ToolingTokensRotate(a.TokenGetRefresh())
	if err != nil {
		log.WithField("err", err).Error("failed to rotate token")

		a.LoadTokensFromEnv()
		retry_limit := 3

		for {
			time.Sleep(time.Second)

			if token, err = a.ToolingTokensRotate(a.TokenGetRefresh()); err == nil {
				break
			}

			log.WithField("err", err).Error("failed to rotate token")

			retry_limit--

			if retry_limit < 1 {
				os.Exit(1)
			}
		}
	}

	a.TokenSetAccess(token.Token)
	a.TokenSetExpirationTime(token.Exp)
	a.TokenSetRefresh(token.RefreshToken)

	log.Info("saving new token")
	if err := a.Save(); err != nil {
		log.WithField("err", err).Error()

		return
	}

	a.SlackClient = a.factory(a.TokenGetAccess())
}

func (a *App) Run(ctx context.Context, duration time.Duration) {
	log.Info("launch the ticker")

	t := time.NewTicker(duration)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			log.Info("ticker launch of the check")
			a.check()
		}
	}
}

func (a *App) check() {
	a.SlackClient = a.factory(a.TokenGetAccess())

	log.Info("checking access token")
	if token, err := a.AuthTest(); err != nil {
		log.WithField("err", err).Error("failed to verify current token")
		a.tokenRotate()
	} else {
		log.Debugf("%#v", token)
	}
}

func NewSlack(storage Storage, factory SlackClientFactory) *App {
	a := &App{
		factory: factory,
		Storage: storage,
	}

	log.Info("initial launch of the check")
	a.check()

	return a
}

func New(storageType string) Storage {
	var s Storage

	switch storageType {
	case "awssecrets":
		s = awssecrets.New()
	case "fs":
		s = fs.New()
	case "vault":
		s = vault.New()
	}

	if err := s.Read(); err != nil {
		log.WithField("err", err).Error("reading was failed")
		s.LoadTokensFromEnv()
	}

	return s
}
