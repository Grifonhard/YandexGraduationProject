package auth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type tokenDetails struct {
	AccessToken  string
	AccessUUID   string
	AtExpires    int64
}

const EXPIRED_AT = time.Hour

const (
	AUTHORIZED_KEY = "authorized"
	ACCESS_UUID_KEY = "access_uuid"
	USER_ID_KEY = "user_id"
	USER_NAME_KEY = "user_name"
	CLIENT_IP = "client_ip" 		// для того чтобы гарантировать, что токен используется конкретным устройством, на котором произошёл логин
	CLIENT_MAC = "client_mac"		// для того чтобы гарантировать, что токен используется конкретным устройством, на котором произошёл логин
	EXPIRED_KEY = "exp"
)

var accessSecret = []byte("very-very-secret-key-for-tokens")

func createToken(userid int, username string, tTLAccess time.Duration) (*tokenDetails, error) {
	td := &tokenDetails{}
	td.AtExpires = time.Now().Add(tTLAccess).Unix()
	// uuid длинной 36 символов
	td.AccessUUID = uuid.New().String()
	
	var err error
	// Создание Access Token
	atClaims := jwt.MapClaims{}
	atClaims[AUTHORIZED_KEY] = true
	atClaims[ACCESS_UUID_KEY] = td.AccessUUID
	atClaims[USER_ID_KEY] = userid
	atClaims[USER_NAME_KEY] = username
	atClaims[EXPIRED_KEY] = td.AtExpires
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

func getUserFromClaims(claims jwt.MapClaims) (*User, error) {
	var user User

	authBoolInter, ok := claims[AUTHORIZED_KEY]
	if !ok {
		return nil, ErrBadClaims
	}

	uuidInter, ok := claims[ACCESS_UUID_KEY]
	if !ok {
		return nil, ErrBadClaims
	}

	userIDInter, ok := claims[USER_ID_KEY]
	if !ok {
		return nil, ErrBadClaims
	}

	userNameInter, ok := claims[USER_NAME_KEY]
	if !ok {
		return nil, ErrBadClaims
	}

	expiratedAtInter, ok := claims[EXPIRED_KEY]
	if !ok {
		return nil, ErrBadClaims
	}

	authBool, ok := authBoolInter.(bool)
	if !ok {
		return nil, ErrBadClaims
	}

	user.IsAutorized = authBool

	uuid, ok := uuidInter.(string)
	if !ok {
		return nil, ErrBadClaims
	}

	user.TokenUUID = uuid

	userID, ok := userIDInter.(int)
	if !ok {
		return nil, ErrBadClaims
	}

	user.UserID = userID

	username, ok := userNameInter.(string)
	if !ok {
		return nil, ErrBadClaims
	}

	user.Username = username

	expInt64, ok := expiratedAtInter.(int64)
	if !ok {
		return nil, ErrBadClaims
	}

	user.ExpiredAt = time.Unix(expInt64, 0)

	return &user, nil
}

// checkClient проверяет что этот тот же клиент по мак адресу и ip 
func checkClient(tokenString string, mac, ip string) error {
	claims, err := decodeToken(tokenString)
	if err != nil {
		return fmt.Errorf("decode token fail %w", err)
	}

	if claims[CLIENT_MAC] != mac || claims[CLIENT_IP] != ip {
		return ErrBadClient
	} else {
		return nil
	}
}