package auth

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
)

type JWTService struct {
	// TODO: use token repo to manage invalidated token
	tokenRepository repository.TokenRepository
	accessSecret    []byte
	refreshSecret   []byte
	userValidToken  map[string]bool
}

type AccessDetails struct {
	AccessUuid string
	UserId     uint64
}

// userValidToken := map[string]bool

// NewJWTService create message service from repository
func NewJWTService(tokenRepo repository.TokenRepository, accessSecret []byte, refreshSecret []byte, userValidToken map[string]bool) *JWTService {
	return &JWTService{
		tokenRepository: tokenRepo,
		accessSecret:    accessSecret,
		refreshSecret:   refreshSecret,
		userValidToken:  userValidToken,
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
		return model.JWTClaim{}, err
	}

	if !jwtToken.Valid {
		return model.JWTClaim{}, errors.New("token is invalid" + jwtToken.Claims.Valid().Error())
	}

	return claim, nil
}

func (service *JWTService) TokenValid(tokenIn string) error {
	token, err := service.VerifyToken(tokenIn)
	if err != nil {
		return err
	}
	newToken, err := mapClaimToModel(token)

	log.Println(newToken)

	if newToken.Authorized == false /* && newToken.jwt.StandardClaims.Valid*/ {
		return err
	}
	return nil
}

func (service *JWTService) ExtractTokenMetadata(tokenIn string) (*AccessDetails, error) {
	token, err := service.VerifyToken(tokenIn)
	if err != nil {
		return nil, err
	}
	claims := token

	accessUuid := claims.AccessUuid
	userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims.UserId), 10, 64)
	if err != nil {
		return nil, err
	}
	return &AccessDetails{
		AccessUuid: accessUuid,
		UserId:     userId,
	}, nil

	return nil, err
}

func (service *JWTService) DeleteAuth(givenUuid string) error {
	_, notInMap := service.userValidToken[givenUuid]
	if !notInMap {
		return errors.New("dont have value uuid")
	}
	service.userValidToken[givenUuid] = false
	return nil
}

func mapClaimToModel(token model.JWTClaim) (model.JWTClaim, error) {
	tokenMap := model.JWTClaim{}
	// convert json to struct
	jsonString, err := json.Marshal(token)
	if err != nil {
		return tokenMap, err
	}
	json.Unmarshal(jsonString, &tokenMap)
	// tokenRole,  := base64.StdEncoding.DecodeString(tokenMap.Role)
	// tokenUserId, _ := base64.StdEncoding.DecodeString(tokenMap.UserId)

	// tokenMap.Role = byteToString(tokenRole)
	// tokenMap.UserId = byteToString(tokenUserId)

	return tokenMap, err
}
