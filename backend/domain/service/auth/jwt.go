package auth

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
)

type JWTService struct {
	// TODO: use token repo to manage invalidated token
	tokenRepository repository.TokenRepository
	accessSecret    []byte
	refreshSecret   []byte
}

// NewJWTService create message service from repository
func NewJWTService(tokenRepo repository.TokenRepository, accessSecret []byte, refreshSecret []byte) *JWTService {
	return &JWTService{
		tokenRepository: tokenRepo,
		accessSecret:    accessSecret,
		refreshSecret:   refreshSecret,
	}
}

var (
	ACCESSTOKENEXPIRES = 9999999 //Min // I think there should be no expire ?
	REFRESHTOKENSECRET = 24 * 7  //Hours
)

// CreateToken create JWTToken from provied userDetail (userID + role)
func (service *JWTService) CreateToken(userDetail model.UserDetail) (*model.TokenDetails, error) {
	td := &model.TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * time.Duration(ACCESSTOKENEXPIRES)).Unix()
	td.AccessUuid = uuid.NewV4().String()
	td.RtExpires = time.Now().Add(time.Hour * time.Duration(REFRESHTOKENSECRET)).Unix()
	td.RefreshUuid = uuid.NewV4().String()

	var err error
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userDetail.UserId
	atClaims["role"] = userDetail.Role
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString(service.accessSecret)
	if err != nil {
		return nil, err
	}

	// Claims Refresh token payload map
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userDetail.UserId
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString(service.refreshSecret)
	if err != nil {
		return nil, err
	}
	return td, nil
}

// VerifyToken verify token string and return claims (the payload)
func (service *JWTService) VerifyToken(token string) (model.JWTClaim, error) {
	// token extracted will be stored in this struct directly
	var claim model.JWTClaim

	jwtToken, err := jwt.ParseWithClaims(token, &claim, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return service.accessSecret, nil
	})
	if err != nil {
		return model.JWTClaim{}, fmt.Errorf("token is invalid: %w", err)
	}
	if !jwtToken.Valid {
		return model.JWTClaim{}, fmt.Errorf("token is invalid: %w", jwtToken.Claims.Valid())
	}

	if cnt, err := service.tokenRepository.CountToken(model.TokenDBFilter{
		Token: token,
	}); err != nil {
		return model.JWTClaim{}, fmt.Errorf("can't check token: %w", err)
	} else if cnt != 0 {
		return model.JWTClaim{}, errors.New("token is invalid")
	}

	return claim, nil
}

// DeleteToken invalidate the token
func (service *JWTService) InvalidateToken(token string) error {
	// token extracted will be stored in this struct directly
	var claim model.JWTClaim

	jwtToken, err := jwt.ParseWithClaims(token, &claim, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return service.accessSecret, nil
	})
	if err != nil {
		return fmt.Errorf("token is invalid: %w", err)
	}
	if !jwtToken.Valid {
		return fmt.Errorf("token is invalid: %w", jwtToken.Claims.Valid())
	}

	if err := service.tokenRepository.InsertToken(model.TokenDB{
		Token: token,
	}); err != nil {
		return fmt.Errorf("error inserting invalidated token: %w", err)
	}

	return nil
}
