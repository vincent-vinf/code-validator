package jwtx

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vincent-vinf/code-validator/pkg/util/db"
	"log"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
)

const (
	IdentityKey = "id"
	appRealm    = "code-validator"
)

var (
	authMiddleware *jwt.GinJWTMiddleware
)

type loginForm struct {
	Email  string `form:"email" json:"email" binding:"required"`
	Passwd string `form:"passwd" json:"passwd" binding:"required"`
}

type RegisterForm struct {
	Username string `form:"username" json:"username" binding:"required"`
	Email    string `form:"email" json:"email" binding:"required"`
	Passwd   string `form:"passwd" json:"passwd" binding:"required"`
}

type TokenUserInfo struct {
	ID int
}

func GetAuthMiddleware(secret string, timeout, maxRefresh time.Duration) (*jwt.GinJWTMiddleware, error) {
	var err error

	authMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm:       appRealm,
		Key:         []byte(secret),
		Timeout:     timeout,
		MaxRefresh:  maxRefresh,
		IdentityKey: IdentityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*TokenUserInfo); ok {
				return jwt.MapClaims{
					IdentityKey: v.ID,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &TokenUserInfo{
				ID: claims[IdentityKey].(int),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginInfo loginForm
			if err := c.ShouldBind(&loginInfo); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			email := loginInfo.Email
			passwd := loginInfo.Passwd

			if email == "" || passwd == "" {
				return nil, jwt.ErrFailedAuthentication
			}
			id, err := db.Login(email, passwd)
			if err != nil {
				log.Println(err)
				return nil, jwt.ErrFailedAuthentication
			}
			u := &TokenUserInfo{
				ID: id,
			}
			return u, nil
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})
	if err != nil {
		return nil, err
	}
	err = authMiddleware.MiddlewareInit()
	if err != nil {
		return nil, errors.New("authMiddleware.MiddlewareInit() Error:" + err.Error())
	}
	return authMiddleware, nil
}

func GenerateToken(id int) string {
	token, _, err := authMiddleware.TokenGenerator(
		&TokenUserInfo{
			ID: id,
		},
	)
	if err != nil {
		return ""
	}
	fmt.Printf("token (%v, %T)\n", token, token)

	return token
}
