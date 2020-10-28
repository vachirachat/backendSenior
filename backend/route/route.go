package route

import (
	"backendSenior/api"
	"backendSenior/repository"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
)

func NewRouteProduct(route *gin.Engine, connectionDB *mgo.Session) {
	productRepository := repository.ProductRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	productAPI := api.ProductAPI{
		ProductRepository: &productRepository,
	}
	route.GET("api/v1/product", productAPI.ProductListHandler)
	route.POST("api/v1/product", productAPI.AddProductHandeler)
	route.PUT("api/v1/product/:product_id", productAPI.EditProducNametHandler)
	route.DELETE("api/v1/product/:product_id", productAPI.DeleteProductByIDHandler)

	userRepository := repository.UserRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	userAPI := api.UserAPI{
		UserRepository: &userRepository,
	}
	route.GET("api/v1/user", userAPI.UserListHandler)
	route.POST("api/v1/user", userAPI.AddUserHandeler)
	route.PUT("api/v1/user/:user_id", userAPI.EditUserNameHandler)
	route.DELETE("api/v1/user/:user_id", userAPI.DeleteUserByIDHandler)
}
