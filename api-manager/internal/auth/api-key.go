package auth

import (
	"context"
	"time"

	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/shared/amqp"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
)

type APIKeyMeta struct {
	// Id is the API Key id.
	Id int32
	// UserId is the user identifier.
	UserId int32
	// Role is the user role.
	Role roles.Role
}

func (a *APIKeyMeta) Bytes() []byte {
	b, _ := amqp.Encode(a)
	return b
}

func (a *Auth) ReadAPIKeyMetadata(bytes []byte) (metadata APIKeyMeta, err error) {
	return metadata, amqp.Decode(bytes, &metadata)
}

// Craetes a API key for a user and saves on Redis.
func (a *Auth) NewAPIKey(ctx context.Context, apikey APIKeyMeta, duration time.Duration) (string, error) {
	token, err := NewToken(a.TokenSize)
	if err != nil {
		return "", err
	}
	p := a.rdb.Pipeline()
	p.Set(ctx, rdb.AuthReverseAPIKeyKey(apikey.Id), token, duration)
	p.Set(ctx, rdb.AuthAPIKeyKey(token), apikey.Bytes(), duration)
	_, err = p.Exec(ctx)
	if err != nil {
		return "", err
	}
	return token, nil
}

// RemoveAPIKey removes a user API key. If API key doesn't exists returns an error.
func (a *Auth) RemoveAPIKey(ctx context.Context, id int32) error {
	apikey, err := a.rdb.Get(ctx, rdb.AuthReverseAPIKeyKey(id)).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	return a.rdb.Del(ctx, rdb.AuthAPIKeyKey(apikey), rdb.AuthReverseAPIKeyKey(id)).Err()
}

// ValidateAPIKey validates a API key on Redis and return the key metadata. An error
// is returned if fail to comunicate with Redis or API key doesn't exist.
func (a *Auth) ValidateAPIKey(ctx context.Context, apikey string) (metadata APIKeyMeta, err error) {
	c := a.rdb.Get(ctx, rdb.AuthAPIKeyKey(apikey))
	b, err := c.Bytes()
	if err != nil {
		return metadata, err
	}
	metadata, err = a.ReadAPIKeyMetadata(b)
	if err != nil {
		return metadata, err
	}
	return metadata, nil
}
