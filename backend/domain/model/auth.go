package model

import (
	"github.com/dgrijalva/jwt-go"
)

type Permission struct {
	Resource string   `json:"resource" bson:"resource"`
	Scopes   []string `json:"scopes" bson:"scopes"`
}

// TokenDetails represent pair of access token and refresh token along with other informations
type TokenDetails struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	AccessUuid   string `json:"-"`
	RefreshUuid  string `json:"-"`
	AtExpires    int64  `json:"expiresAt"`
	RtExpires    int64  `json:"refreshExpiresAt"`
}

type AccessDetails struct {
	AccessUuid string
	UserId     uint64
}

// UserDetail is detail of user to be used to create JWTClaim.
// It's also used for checking permission
type UserDetail struct {
	Role   string `json:"role"`
	UserId string `json:"user_id"`
}

// JWTClaim is the payload of JWT
type JWTClaim struct {
	AccessUuid string `json:"access_uuid"`
	Authorized bool   `json:"authorized"`
	Role       string `json:"role"`
	UserId     string `json:"user_id"`
	// Inherit standard claims
	jwt.StandardClaims
}

// JWTClaim is the payload of JWT
type TokenDB struct {
	Token string `bson:"_id,omitempty"`
}

type TokenDBFilter struct {
	Token interface{} `bson:"_id,omitempty"`
}
