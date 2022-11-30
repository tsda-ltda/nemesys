package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/fernandotsda/nemesys/shared/rdb"
	"github.com/go-redis/redis/v8"
)

const SessionCookieName = "sess"

type SessionMeta struct {
	UserId int32
	Role   roles.Role
}

func (m *SessionMeta) Bytes() []byte {
	// join userId and role into a single string
	return []byte(fmt.Sprintf("%d=%d", m.UserId, m.Role))
}

func (a *Auth) ReadSessionMetadata(bytes []byte) (metadata SessionMeta, err error) {
	splited := strings.Split(string(bytes), "=")
	userId, err := strconv.ParseInt(splited[0], 10, 32)
	if err != nil {
		return metadata, err
	}
	metadata.UserId = int32(userId)
	role, err := strconv.Atoi(splited[1])
	if err != nil {
		return metadata, err
	}
	metadata.Role = uint8(role)
	return metadata, nil
}

// Crates a new session for a user and saves on Redis and remove any old session.
func (a *Auth) NewSession(ctx context.Context, meta SessionMeta) (string, error) {
	err := a.RemoveSession(ctx, meta.UserId)
	if err != nil {
		return "", err
	}

	token, err := NewToken(a.TokenSize)
	if err != nil {
		return "", err
	}

	p := a.rdb.Pipeline()
	p.Set(ctx, rdb.AuthReverseSessionKey(meta.UserId), token, a.SessionTTL)
	p.Set(ctx, rdb.AuthSessionKey(token), meta.Bytes(), a.SessionTTL)
	_, err = p.Exec(ctx)
	if err != nil {
		return "", err
	}
	return token, nil
}

// RemoveSession removes a user session. If session doesn't exists returns an error.
func (a *Auth) RemoveSession(ctx context.Context, userId int32) error {
	oldToken, err := a.rdb.Get(ctx, rdb.AuthReverseSessionKey(userId)).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	return a.rdb.Del(ctx, rdb.AuthSessionKey(oldToken), rdb.AuthReverseSessionKey(userId)).Err()
}

// Validate validates a session on Redis and return the user metadata. An error
// is returned if fail to comunicate with Redis or session doesn't exist.
func (a *Auth) Validate(ctx context.Context, session string) (metadata SessionMeta, err error) {
	c := a.rdb.Get(ctx, rdb.AuthSessionKey(session))
	b, err := c.Bytes()
	if err != nil {
		return metadata, err
	}
	metadata, err = a.ReadSessionMetadata(b)
	if err != nil {
		return metadata, err
	}
	return metadata, nil
}
