package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/fernandotsda/nemesys/api-manager/internal/roles"
	"github.com/go-redis/redis/v8"
)

type SessionMeta struct {
	UserId int
	Role   roles.Role
}

func (m *SessionMeta) Bytes() []byte {
	// join userId and role into a single string
	return []byte(fmt.Sprintf("%d=%d", m.UserId, m.Role))
}

func (a *Auth) ReadSessionMetadata(bytes []byte) (metadata SessionMeta, err error) {
	// split string
	splited := strings.Split(string(bytes), "=")

	// get userId
	userId, err := strconv.Atoi(splited[0])
	if err != nil {
		return metadata, err
	}
	metadata.UserId = userId

	// get role
	role, err := strconv.Atoi(splited[1])
	if err != nil {
		return metadata, err
	}
	metadata.Role = uint8(role)

	return metadata, nil
}

// Crates a new session for a user and saves on Redis and remove any old session.
func (a *Auth) NewSession(ctx context.Context, meta SessionMeta) (string, error) {
	// get old session
	oldToken, err := a.rdb.Get(ctx, ReverseSessionKey(meta.UserId)).Result()
	if err != nil && err != redis.Nil {
		return "", err
	}

	// generate token
	token, err := NewToken(a.TokenSize)
	if err != nil {
		return "", err
	}

	// create command pipeline
	p := a.rdb.Pipeline()

	// delete session and reverse session key
	p.Del(ctx, SessionKey(oldToken), ReverseSessionKey(meta.UserId))

	// save reverse session key
	p.Set(ctx, ReverseSessionKey(meta.UserId), token, a.SessionTTL)

	// save session
	p.Set(ctx, SessionKey(token), meta.Bytes(), a.SessionTTL)

	// exec pipeline
	_, err = p.Exec(ctx)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Validate validates a session on Redis and return the user metadata. An error
// is returned if fail to comunicate with Redis or session doesn't exist.
func (a *Auth) Validate(ctx context.Context, session string) (metadata SessionMeta, err error) {
	// Check rdb
	c := a.rdb.Get(ctx, SessionKey(session))
	b, err := c.Bytes()
	if err != nil {
		return metadata, err
	}

	// Read metada
	metadata, err = a.ReadSessionMetadata(b)
	if err != nil {
		return metadata, err
	}

	return metadata, nil
}