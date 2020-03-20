// Copyright 2019 GoAdmin Core Team. All rights reserved.
// Use of this source code is governed by a Apache-2.0 style
// license that can be found in the LICENSE file.

package auth

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/service"
	"github.com/GoAdminGroup/go-admin/plugins/admin/models"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

// Auth get the user model from Context.
func Auth(ctx *context.Context) models.UserModel {
	return ctx.User().(models.UserModel)
}

// Check check the password and username and return the user model.
func Check(password string, username string, conn db.Connection) (user models.UserModel, ok bool) {

	user = models.User().SetConn(conn).FindByUserName(username)

	if user.IsEmpty() {
		ok = false
	} else {
		if comparePassword(password, user.Password) {
			ok = true
			user = user.WithRoles().WithPermissions().WithMenus()
			user.UpdatePwd(EncodePassword([]byte(password)))
		} else {
			ok = false
		}
	}
	return
}

func comparePassword(comPwd, pwdHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(pwdHash), []byte(comPwd))
	return err == nil
}

// EncodePassword encode the password.
func EncodePassword(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash[:])
}

// SetCookie set the cookie.
func SetCookie(ctx *context.Context, user models.UserModel, conn db.Connection) bool {
	InitSession(ctx, conn).Add("user_id", user.Id)
	return true
}

// DelCookie delete the cookie from Context.
func DelCookie(ctx *context.Context, conn db.Connection) bool {
	InitSession(ctx, conn).Clear()
	return true
}

type TokenService struct {
	tokens CSRFToken
	lock   sync.Mutex
}

func (s *TokenService) Name() string {
	return TokenServiceKey
}

func init() {
	service.Register(TokenServiceKey, func() (service.Service, error) {
		return &TokenService{
			tokens: make(CSRFToken, 0),
		}, nil
	})
}

const (
	TokenServiceKey = "token_csrf_helper"
	ServiceKey      = "auth"
)

func GetTokenService(s interface{}) *TokenService {
	if srv, ok := s.(*TokenService); ok {
		return srv
	}
	panic("wrong service")
}

// AddToken add the token to the CSRFToken.
func (s *TokenService) AddToken() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	tokenStr := modules.Uuid()
	s.tokens = append(s.tokens, tokenStr)
	return tokenStr
}

// CheckToken check the given token with tokens in the CSRFToken, if exist
// return true.
func (s *TokenService) CheckToken(toCheckToken string) bool {
	for i := 0; i < len(s.tokens); i++ {
		if (s.tokens)[i] == toCheckToken {
			s.tokens = append((s.tokens)[:i], (s.tokens)[i+1:]...)
			return true
		}
	}
	return false
}

// CSRFToken is type of a csrf token list.
type CSRFToken []string

type Processor func(ctx *context.Context) (model models.UserModel, exist bool, msg string)

type Service struct {
	P Processor
}

func (s *Service) Name() string {
	return "auth"
}

func GetService(s interface{}) *Service {
	if srv, ok := s.(*Service); ok {
		return srv
	}
	panic("wrong service")
}

func NewService(processor Processor) *Service {
	return &Service{
		P: processor,
	}
}
