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
	err = service.tokenRepository.AddToken(userDetail.UserId, td.AccessToken)
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

	tokenDB, err := service.tokenRepository.VerifyDBToken(claim.UserId, token)
	if err != nil || token != tokenDB {
		return model.JWTClaim{}, errors.New("token is invalid")
	}

	if err != nil {
		return model.JWTClaim{}, err
	}

	if !jwtToken.Valid {
		return model.JWTClaim{}, errors.New("token is invalid" + jwtToken.Claims.Valid().Error())
	}

	return claim, nil
}

// DeleteToken verify token string and return error
func (service *JWTService) RemoveToken(userid string) error {
	err := service.tokenRepository.RemoveToken(userid)
	if err != nil {
		return err
	}
	return nil
}

// DeleteToken verify token string and return error
func (service *JWTService) GetAllToken() ([]model.TokenDB, error) {
	// token extracted will be stored in this struct directly
	var tokens []model.TokenDB

	tokens, err := service.tokenRepository.GetAllToken()
	if err != nil {
		return []model.TokenDB{}, err
	}

	return tokens, nil
}
