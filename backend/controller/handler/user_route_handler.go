package route

import (
	authMw "backendSenior/controller/middleware/auth"
	"backendSenior/domain/service"
	"backendSenior/domain/service/auth"
	"backendSenior/utills"
	"fmt"
	"os"

	// "strconv"

	"backendSenior/domain/model"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

// UserRouteHandler is handler for route
type UserRouteHandler struct {
	userService    *service.UserService
	jwtService     *auth.JWTService
	authMiddleware *authMw.JWTMiddleware
}

func NewUserRouteHandler(userService *service.UserService, jwtService *auth.JWTService, authMiddleware *authMw.JWTMiddleware) *UserRouteHandler {
	return &UserRouteHandler{
		userService:    userService,
		jwtService:     jwtService,
		authMiddleware: authMiddleware,
	}
}

// Mount make handle handle request for specified routerGroup
func (handler *UserRouteHandler) Mount(routerGroup *gin.RouterGroup) {
	routerGroup.GET("user", handler.userListHandler)
	routerGroup.PUT("user/updateuserprofile", handler.editUserNameHandler)
	routerGroup.DELETE("byid/:user_id", handler.deleteUserByIDHandler)
	routerGroup.POST("getuserbyemail", handler.getUserByEmail)

	//SignIN/UP API
	// routerGroup.GET("/token", handler.userTokenListHandler)
	routerGroup.POST("/login", handler.loginHandle)
	routerGroup.POST("/logout", handler.logoutHandle)
	routerGroup.POST("/signup", handler.addUserSignUpHandeler)
	routerGroup.POST("/token/refresh", handler.Refresh)
	routerGroup.GET("/me", handler.authMiddleware.AuthRequired(), handler.getMeHandler)

	// (for proxy)
	routerGroup.POST("/verify", handler.verifyToken)
	routerGroup.GET("/byid/:id", handler.getUserByIDHandler)
}

func (handler *UserRouteHandler) getMeHandler(context *gin.Context) {
	id := context.GetString(authMw.UserIdField)

	user, err := handler.userService.GetUserByID(id)
	if err != nil {
		log.Println("error GetMe", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, user)
}

func (handler *UserRouteHandler) userListHandler(context *gin.Context) {
	var usersInfo model.UserInfo
	users, err := handler.userService.GetAllUsers()
	if err != nil {
		log.Println("error userListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	usersInfo.User = users
	context.JSON(http.StatusOK, usersInfo)
}

// for get user by id
func (handler *UserRouteHandler) getUserByIDHandler(context *gin.Context) {
	userID := context.Param("id")
	if !bson.IsObjectIdHex(userID) {
		context.JSON(http.StatusBadRequest, gin.H{"status": "bad user id"})
		return
	}
	user, err := handler.userService.GetUserByID(userID)
	if err != nil {
		log.Println("error GetUserByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	context.JSON(http.StatusOK, user)
}

// GetUserByEmail for get user by id
func (handler *UserRouteHandler) getUserByEmail(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	user, err = handler.userService.GetUserByEmail(user.Email)
	if err != nil {
		log.Println("error GetUserByEmailHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, user)
}

//for return roomidList of User
// func (handler *UserRouteHandler) getUserRoomByUserID(context *gin.Context) {
// 	var user model.User
// 	err := context.ShouldBindJSON(&user)
// 	userResult, err := handler.userService.GetUserByID(user.UserID)
// 	if err != nil {
// 		log.Println("error getUserRoomByUserID", err.Error())
// 		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
// 		return
// 	}
// 	roomIDList := userResult.Room
// 	log.Println(roomIDList)
// 	var roomNameList []string
// 	for _, s := range roomIDList {
// 		room, err := handler.userService.GetRoomWithRoomID(s)
// 		if err != nil {
// 			log.Println("error getUserRoomByUserID", err.Error())
// 			context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
// 			return
// 		}
// 		roomNameList = append(roomNameList, room.RoomName)
// 	}

// 	context.JSON(http.StatusOK, gin.H{"username": userResult.Name, "RoomIDList": userResult.Room, "RoomNameList": roomNameList})
// }

// AddUserHandeler api
func (handler *UserRouteHandler) addUserHandeler(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error AddUserHandeler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}
	err = handler.userService.AddUser(user)
	if err != nil {
		log.Println("error AddUserHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

// EditUserNameHandler api
func (handler *UserRouteHandler) editUserNameHandler(context *gin.Context) {
	var user model.User
	// userID := context.Param("user_id")
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error EditUserNametHandler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}
	err = handler.userService.EditUserName(user.UserID.Hex(), user)
	if err != nil {
		log.Println("error EditUserNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *UserRouteHandler) updateUserHandler(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error UpdateUserHandler", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}
	log.Println(user.UserID)
	err = handler.userService.UpdateUser(user.UserID.Hex(), user)
	if err != nil {
		log.Println("error UpdateUserHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (handler *UserRouteHandler) deleteUserByIDHandler(context *gin.Context) {
	userID := context.Param("user_id")
	err := handler.userService.DeleteUserByID(userID)
	if err != nil {
		log.Println("error DeleteUserHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
	}
	context.JSON(http.StatusNoContent, gin.H{"status": "success"})
}

// // Edit user role
// func (handler *UserRouteHandler) editUseRoleHandler(context *gin.Context) {
// 	var credentials model.UserSecret
// 	err := context.ShouldBindJSON(&credentials)
// 	if err != nil {
// 		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
// 		return
// 	}
// 	err = handler.userService.EditUserRole(credentials)
// 	if err != nil {
// 		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
// 		return
// 	}
// 	context.JSON(http.StatusOK, gin.H{"status": "success"})
// }

func (handler *UserRouteHandler) loginHandle(context *gin.Context) {
	var credentials model.UserSecret
	err := context.ShouldBindJSON(&credentials)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	user, err := handler.userService.Login(credentials.Email, credentials.Password)
	tokenDetails, err := handler.jwtService.CreateToken(model.UserDetail{
		Role:   utills.ROLEUSER, // TODO: placeholder, implement role later
		UserId: user.UserID.Hex(),
	})
	log.Println(tokenDetails)

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"status": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success", "token": tokenDetails})
}

func (handler *UserRouteHandler) logoutHandle(context *gin.Context) {
	var body struct {
		Token string `json:"token"`
	}
	err := context.ShouldBindJSON(&body)
	log.Println(body.Token)
	metadata, err := handler.jwtService.ExtractTokenMetadata(body.Token)
	if err != nil {
		log.Println("metadata")
		log.Println(metadata)
		log.Println(err)
		context.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	log.Println("metadata")
	log.Println(metadata)
	delErr := handler.jwtService.DeleteAuth(metadata.UserId)
	if delErr != nil {
		context.JSON(http.StatusUnauthorized, delErr.Error())
		return
	}
	context.JSON(http.StatusOK, "Successfully logged out")
}

// Signup API
func (handler *UserRouteHandler) addUserSignUpHandeler(context *gin.Context) {
	var user model.User
	err := context.ShouldBindJSON(&user)
	if err != nil {
		log.Println("error AddUserSignUpHandeler ShouldBindJSON", err.Error())
		context.JSON(http.StatusBadRequest, gin.H{"status": err.Error()})
		return
	}
	err = handler.userService.Signup(user)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"status": err.Error(),
		})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (handler *UserRouteHandler) verifyToken(context *gin.Context) {
	var body struct {
		Token string `json:"token"`
	}
	err := context.ShouldBindJSON(&body)
	log.Println(body.Token)
	if err != nil || body.Token == "" {
		context.JSON(http.StatusBadRequest, gin.H{"status": "error"})
		return
	}

	claim, err := handler.jwtService.VerifyToken(body.Token)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"status": "verify error: " + err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"userId": claim.UserId,
	})
}

func (handler *UserRouteHandler) Refresh(context *gin.Context) {
	mapToken := map[string]string{}
	if err := context.ShouldBindJSON(&mapToken); err != nil {
		context.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	refreshToken := mapToken["refresh_token"]

	//verify the token
	os.Setenv("REFRESH_SECRET", "mcmvmkmsdnfsdmfdsjf") //this should be in an env file
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	//if there is an error, the token must have expired
	if err != nil {
		context.JSON(http.StatusUnauthorized, "Refresh token expired")
		return
	}
	//is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		context.JSON(http.StatusUnauthorized, err)
		return
	}
	//Since token is valid, get the uuid:
	claims, ok := token.Claims.(jwt.MapClaims) //the token claims should conform to MapClaims
	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string) //convert the interface to string
		if !ok {
			context.JSON(http.StatusUnprocessableEntity, err)
			return
		}
		// userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		userId := claims["user_id"]
		if err != nil {
			context.JSON(http.StatusUnprocessableEntity, "Error occurred")
			return
		}
		//Delete the previous Refresh Token
		delErr := handler.jwtService.DeleteAuth(refreshUuid)
		if delErr != nil /*|| deleted == 0*/ { //if any goes wrong
			context.JSON(http.StatusUnauthorized, "unauthorized")
			return
		}
		//Create new pairs of refresh and access tokens
		ts, createErr := handler.jwtService.CreateToken(userId.(model.UserDetail))
		if createErr != nil {
			context.JSON(http.StatusForbidden, createErr.Error())
			return
		}
		//save the tokens metadata to redis
		//didnt save data to redis
		// saveErr := handler.jwtService.CreateAuth(userId, ts)
		// if saveErr != nil {
		// 	context.JSON(http.StatusForbidden, saveErr.Error())
		// 	return
		// }
		tokens := map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}
		context.JSON(http.StatusCreated, tokens)
	} else {
		context.JSON(http.StatusUnauthorized, "refresh expired")
	}
}
