package auth

import (
	"backendSenior/domain/interface/repository"
	"backendSenior/domain/model"
	"backendSenior/utills"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type TokenService struct {
	tokenRepository repository.TokenRepository
}

// NewMessageService create message service from repository
func NewTokenService(tokenRepo repository.TokenRepository) *TokenService {
	return &TokenService{
		tokenRepository: tokenRepo,
	}
}

// Must use From .env file
const (
	ACCESSSECRET  = "Secret"
	REFRESHSECRET = "Secret"
)

var (
	ACCESSTOKENEXPIRES = 15     //Min
	REFRESHTOKENSECRET = 24 * 7 //Hours
)

func CreateToken(userid string) (*model.TokenDetails, error) {
	td := &model.TokenDetails{}
	// Create Access Key expire
	td.AtExpires = time.Now().Add(time.Minute * time.Duration(ACCESSTOKENEXPIRES)).Unix()
	td.AccessUuid = uuid.NewV4().String()
	// Create Refresh Key expire
	td.RtExpires = time.Now().Add(time.Hour * time.Duration(REFRESHTOKENSECRET)).Unix()
	td.RefreshUuid = uuid.NewV4().String()

	// Encode payload  userid / roloeUser
	userIdb64 := base64.StdEncoding.EncodeToString([]byte(userid))
	roleUserb64 := base64.StdEncoding.EncodeToString([]byte(utills.ROLEUSER))

	var err error
	// Claims Access token payload map
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userIdb64
	atClaims["role"] = roleUserb64
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	// Sign with SigningMethodHS256
	td.AccessToken, err = at.SignedString([]byte(ACCESSSECRET))
	if err != nil {
		return nil, err
	}

	// Claims Refresh token payload map
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userid
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(REFRESHSECRET))
	if err != nil {
		return nil, err
	}
	return td, nil
}

func VerifyToken(context *gin.Context) (*jwt.Token, error) {
	tokenString := extractToken(context)
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(ACCESSSECRET), nil
	})

	if err != nil {
		context.Abort()
		context.Writer.WriteHeader(http.StatusUnauthorized)
		context.Writer.Write([]byte("Unauthorized: not login state"))
		return nil, err
	}
	return token, nil
}

// Remap interface MapClaim to Struct
func mapClaimToModel(token *jwt.Token) model.JWTClaim {
	jsonString, _ := json.Marshal(token.Claims)
	// convert json to struct
	tokenMap := model.JWTClaim{}
	json.Unmarshal(jsonString, &tokenMap)
	tokenRole, _ := base64.StdEncoding.DecodeString(tokenMap.Role)
	tokenUserId, _ := base64.StdEncoding.DecodeString(tokenMap.UserId)

	tokenMap.Role = byteToString(tokenRole)
	tokenMap.UserId = byteToString(tokenUserId)

	return tokenMap
}

func byteToString(byteArr []byte) string {
	return string(byteArr[:])
}

func extractToken(context *gin.Context) string {
	// Get JWT from Header Authorization
	bearToken := context.Request.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

// SAVE to DB
// func CreateAuth(userid uint64, td *model.TokenDetails) error {
// 	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
// 	rt := time.Unix(td.RtExpires, 0)
// 	now := time.Now()
// 	Write to DB Anywhere
// 	errAccess := client.Set(td.AccessUuid, strconv.Itoa(int(userid)), at.Sub(now)).Err()
// 	if errAccess != nil {
// 		return errAccess
// 	}
// 	errRefresh := client.Set(td.RefreshUuid, strconv.Itoa(int(userid)), rt.Sub(now)).Err()
// 	if errRefresh != nil {
// 		return errRefresh
// 	}
// 	return nil
// }

// func TokenValid(context *gin.Context) error {
// 	token, err := verifyToken(context)
// 	if err != nil {

// 		return err
// 	}
// 	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
// 		return err
// 	}
// 	return nil
// }

// func ExtractTokenMetadata(context *gin.Context) (*model.AccessDetails, error) {
// 	token, err := verifyToken(context)
// 	if err != nil {
// 		return nil, err
// 	}
// 	claims, ok := token.Claims.(jwt.MapClaims)
// 	if ok && token.Valid {
// 		accessUuid, ok := claims["access_uuid"].(string)
// 		if !ok {
// 			return nil, err
// 		}
// 		userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return &model.AccessDetails{
// 			AccessUuid: accessUuid,
// 			UserId:     userId,
// 		}, nil
// 	}
// 	return nil, err
// }

// func FetchAuth(authD *AccessDetails) (uint64, error) {
// 	userid, err := client.Get(authD.AccessUuid).Result()
// 	if err != nil {
// 		return 0, err
// 	}
// 	userID, _ := strconv.ParseUint(userid, 10, 64)
// 	return userID, nil
// }
