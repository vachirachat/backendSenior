package api_test

import (
	"./backendSenior/api"
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
)

func Test_ProductListHandler_Should_Be_ProductInfo(t *testing.T) {
	expected := listProduct
	request := httptest.NewRequest("GET", "/api/v1/product", nil)
	writer := httptest.NewRecorder()
	productAPI := api.ProductAPI{
		ProductRepository: &mockProductRepository{},
	}

	testRoute := gin.Default()
	testRoute.GET("api/v1/product", productAPI.ProductListHandler)
	testRoute.ServeHTTP(writer, request)

	response := writer.Result()
	actualRespone, _ := ioutil.ReadAll(response.Body)

	assert.Equal(t, expected, string(actualRespone))
}

func Test_GetProductByIDHandler_Input_Id_5befe40d9c71fe169a4341df_Should_Be_Product_Name_M150(t *testing.T) {
	expected := `{"product_id":"5befe40d9c71fe169a4341df","product_name":"M150","product_price":"14.00","amount":20,"updated_time":"0001-01-01T00:00:00Z"}`
	request := httptest.NewRequest("GET", "/api/v1/product/5befe40d9c71fe169a4341df", nil)
	writer := httptest.NewRecorder()
	productAPI := api.ProductAPI{
		ProductRepository: &mockProductRepository{},
	}

	testRoute := gin.Default()
	testRoute.GET("api/v1/product/:product_id", productAPI.GetProductByIDHandler)
	testRoute.ServeHTTP(writer, request)

	response := writer.Result()
	actual, _ := ioutil.ReadAll(response.Body)

	assert.Equal(t, expected, string(actual))
}

func Test_AddProductHandler_Input_Product_Shoud_Be_Created(t *testing.T) {
	expectedStatusCode := http.StatusCreated
	product := []byte(`{"product_name":"Orio","product_price":"5.00","amount":50}`)
	request := httptest.NewRequest("POST", "/api/v1/product", bytes.NewBuffer(product))
	writer := httptest.NewRecorder()
	productAPI := api.ProductAPI{
		ProductRepository: &mockProductRepository{},
	}

	testRoute := gin.Default()
	testRoute.POST("api/v1/product", productAPI.AddProductHandeler)
	testRoute.ServeHTTP(writer, request)

	response := writer.Result()
	actualStatusCode := response.StatusCode

	assert.Equal(t, expectedStatusCode, actualStatusCode)
}

func Test_EditProducNametHandler_Input_Name_M150_Shoukd_Be_Edited(t *testing.T) {
	expectedStatusCode := http.StatusOK
	product := []byte(`{"product_name":"M150"}`)
	request := httptest.NewRequest("PUT", "/api/v1/product/5beaf7bd62e63844ce22cc58", bytes.NewBuffer(product))
	writer := httptest.NewRecorder()
	productAPI := api.ProductAPI{
		ProductRepository: &mockProductRepository{},
	}

	testRoute := gin.Default()
	testRoute.PUT("api/v1/product/:product_id", productAPI.EditProducNametHandler)
	testRoute.ServeHTTP(writer, request)

	response := writer.Result()
	actualStatusCode := response.StatusCode

	assert.Equal(t, expectedStatusCode, actualStatusCode)
}

func Test_DeleteProductByIDHandler_Input_Id_5befe40d9c71fe169a4341df_Shout_Be_Delete_Product_Name_M150(t *testing.T) {
	expectedStatusCode := http.StatusNoContent
	request := httptest.NewRequest("DELETE", "/api/v1/product/5befe40d9c71fe169a4341df", nil)
	writer := httptest.NewRecorder()
	productAPI := api.ProductAPI{
		ProductRepository: &mockProductRepository{},
	}

	testRoute := gin.Default()
	testRoute.DELETE("api/v1/product/:product_id", productAPI.DeleteProductByIDHandler)
	testRoute.ServeHTTP(writer, request)

	response := writer.Result()
	actualStatusCode := response.StatusCode

	assert.Equal(t, expectedStatusCode, actualStatusCode)
}
