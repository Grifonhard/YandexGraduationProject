package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"time"
)

type tokenDetails struct {
	AccessToken  string
	AccessUUID   string
	AtExpires    int64
}

const EXPIRED_AT = time.Hour

var accessSecret = []byte("very-very-secret-key-for-tokens")

func createToken(userid uint64, username string, tTLAccess time.Duration) (*tokenDetails, error) {
	td := &tokenDetails{}
	td.AtExpires = time.Now().Add(tTLAccess).Unix()
	td.AccessUUID = uuid.New().String()
	
	var err error
	// Создание Access Token
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUUID
	atClaims["user_id"] = userid
	atClaims["user_name"] = username
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString(accessSecret)
	if err != nil {
		return nil, err
	}
	
	return td, nil
}

func decodeToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnexpSignMethod
		}
		return accessSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok && !token.Valid {
		return nil, ErrInvalidMethod
	}

	// Check if token is expired
	if claims["exp"] != nil {
		exp := claims["exp"].(float64)
		if exp < float64(time.Now().Unix()) {
			return nil, ErrTokenExpired
		}
	} else {
		return nil, ErrNoExpTime
	}

	return claims, nil
}