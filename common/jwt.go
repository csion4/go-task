package common

import (
	"github.com/golang-jwt/jwt"
	"time"
)

var jwtKey = []byte("a_secret")

type Claims struct {
	UserId int
	jwt.StandardClaims
}

func ReleaseToken(userId int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "ngxs.site",
			Subject:   "user token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	return tokenString, err
}


func ParseToken(tokenString string) (*jwt.Token,*Claims,error){
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (i interface{}, err error) {
		return jwtKey,nil
	})

	return token,claims,err
}