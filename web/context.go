// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"reflect"

	"github.com/Ronmi/yugi/actions"
	"github.com/gin-contrib/requestid"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

func newC[T any](c *gin.Context, key string) ContextAccessor[T] {
	return ContextAccessor[T]{c, key}
}

func newS[T any](s sessions.Session, key string) SessionAccessor[T] {
	return SessionAccessor[T]{s, key}
}

type Getter[T any] interface {
	Get() T
	Defaults(v T) T
	NZ() bool
	IfZero(v T) T
}

type Accessor[T any] interface {
	Get() T
	Defaults(v T) T
	NZ() bool
	IfZero(v T) T
	Set(T)
	Del()
}

type ContextAccessor[T any] struct {
	*gin.Context
	key string
}

func (a ContextAccessor[T]) Get() (ret T) {
	x, ok := a.Context.Get(a.key)
	if !ok {
		return
	}
	return x.(T)
}

func (a ContextAccessor[T]) Defaults(v T) T {
	x, ok := a.Context.Get(a.key)
	if !ok {
		return v
	}
	return x.(T)
}

func (a ContextAccessor[T]) NZ() bool {
	x, ok := a.Context.Get(a.key)
	if !ok {
		return false
	}
	return !reflect.ValueOf(x).IsZero()
}

func (a ContextAccessor[T]) IfZero(v T) T {
	if a.NZ() {
		return a.Get()
	}
	return v
}

func (a ContextAccessor[T]) Set(v T) {
	a.Context.Set(a.key, v)
}

func (a ContextAccessor[T]) Del() {
	a.Context.Set(a.key, nil)
}

type Context struct {
	*gin.Context
}

func (c Context) DB() *gorm.DB {
	return c.Context.MustGet("db").(*gorm.DB)
}

func (c Context) ReqID() string {
	return requestid.Get(c.Context)
}

func (c Context) Session() Session {
	return Session{sessions.Default(c.Context)}
}

type SessionAccessor[T any] struct {
	sessions.Session
	key string
}

func (a SessionAccessor[T]) Get() (ret T) {
	x := a.Session.Get(a.key)
	if x == nil {
		return
	}
	return x.(T)
}

func (a SessionAccessor[T]) Defaults(v T) T {
	x := a.Session.Get(a.key)
	if x == nil {
		return v
	}
	return x.(T)
}

func (a SessionAccessor[T]) NZ() bool {
	x := a.Session.Get(a.key)
	if x == nil {
		return false
	}
	return !reflect.ValueOf(x).IsZero()
}

func (a SessionAccessor[T]) IfZero(v T) T {
	if a.NZ() {
		return a.Get()
	}
	return v
}

func (a SessionAccessor[T]) Set(v T) {
	a.Session.Set(a.key, v)
}

func (a SessionAccessor[T]) Del() {
	a.Session.Delete(a.key)
}

type Session struct {
	sessions.Session
}

func (s Session) CurrentUser() Accessor[*actions.User] {
	return newS[*actions.User](s.Session, "current_user")
}

func (s Session) LogCurrentUser() extraLogInfo {
	return func(l *zerolog.Event) *zerolog.Event {
		u := s.CurrentUser()
		if !u.NZ() {
			return l.Interface("current_user", nil)
		}
		return l.Str("current_user", u.Get().Name)
	}
}

func (s Session) OauthState() Accessor[string] {
	return newS[string](s.Session, "oauth_state")
}

func (s Session) Authed() Accessor[bool] {
	return newS[bool](s.Session, "authed")
}

func (s Session) NextMove() Accessor[string] {
	return newS[string](s.Session, "next_move")
}

type OTPKey struct {
	*otp.Key
}

func (o OTPKey) GobEncode() ([]byte, error) {
	return []byte(o.Key.URL()), nil
}

func (o *OTPKey) GobDecode(data []byte) error {
	key, err := otp.NewKeyFromURL(string(data))
	if err != nil {
		return err
	}
	o.Key = key
	return nil
}

func (s Session) OTPKey() Accessor[OTPKey] {
	return newS[OTPKey](s.Session, "otp_key")
}

func (s Session) OTPErrorCount() Accessor[int] {
	return newS[int](s.Session, "otp_error_count")
}

func (s Session) LoginNonce() Accessor[string] {
	return newS[string](s.Session, "login_nonce")
}
