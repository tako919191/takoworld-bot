package envconfig

import (
	"context"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/cockroachdb/errors"
	"github.com/joho/godotenv"
)

// RetryLimit はリトライする最大回数を指定します。
const retryLimit = 3

// RetryInterval はリトライの間隔を指定します。
const retryInterval = 1 // 1秒

type Env struct {
	DISCORD_TOKEN      string
	DISCORD_GUILD_ID   string
	DISCORD_CHANNEL_ID string
	SERVER_ADDRESS     string
	SERVER_PORT        string
	RCON_PORT          string
	RCON_PASSWORD      string
}

func NewEnv() (*Env, error) {
	EnvGet := func(ctx context.Context) error {
		operation := func() error {
			err := godotenv.Load(".env")
			// .envファイルがない場合は無視する
			if err != nil && err.Error() != "open .env: no such file or directory" {
				return nil
			}
			return errors.WithStack(err)
		}
		return retryOperation(ctx, func() error { return operation() })
	}
	err := EnvGet(context.Background())
	if err != nil {
		return nil, err
	}

	return &Env{
		DISCORD_TOKEN:      os.Getenv("DISCORD_TOKEN"),
		DISCORD_GUILD_ID:   os.Getenv("DISCORD_GUILD_ID"),
		DISCORD_CHANNEL_ID: os.Getenv("DISCORD_CHANNEL_ID"),
		SERVER_ADDRESS:     os.Getenv("SERVER_ADDRESS"),
		SERVER_PORT:        os.Getenv("SERVER_PORT"),
		RCON_PORT:          os.Getenv("RCON_PORT"),
		RCON_PASSWORD:      os.Getenv("RCON_PASSWORD"),
	}, nil
}

func retryOperation(ctx context.Context, operation func() error) error {
	retryBackoff := backoff.NewExponentialBackOff()
	retryBackoff.MaxElapsedTime = time.Second * retryInterval

	err := backoff.RetryNotify(func() error {
		err := operation()
		if err != nil {
			return err
		}
		err = backoff.Permanent(err)
		return errors.WithStack(err)
	}, retryBackoff, func(err error, duration time.Duration) {
		//slog.WarnContext(ctx, fmt.Sprintf("%v retrying in %v...", err, duration))
	})
	return errors.WithStack(err)
}
