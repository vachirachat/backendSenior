package api

import (
	"backendSenior/model"
	"backendSenior/repository"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductAPI struct {
	ProductRepository repository.ProductRepository
}

func (api ProductAPI) ProductListHandler(context *gin.Context) {
	var productsInfo model.ProductInfo
	products, err := api.ProductRepository.GetAllProduct()
	if err != nil {
		log.Println("error productListHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	productsInfo.Product = products
	context.JSON(http.StatusOK, productsInfo)
}

func (api ProductAPI) GetProductByIDHandler(context *gin.Context) {
	productID := context.Param("product_id")
	product, err := api.ProductRepository.GetProductByID(productID)
	if err != nil {
		log.Println("error GetProductByIDHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	context.JSON(http.StatusOK, product)
}

func (api ProductAPI) AddProductHandeler(context *gin.Context) {
	var product model.Product
	err := context.ShouldBindJSON(&product)
	if err != nil {
		log.Println("error AddProductHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	err = api.ProductRepository.AddProduct(product)
	if err != nil {
		log.Println("error AddProductHandeler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"status": "success"})
}

func (api ProductAPI) EditProducNametHandler(context *gin.Context) {
	var product model.Product
	productID := context.Param("product_id")
	err := context.ShouldBindJSON(&product)
	if err != nil {
		log.Println("error EditProducNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	err = api.ProductRepository.EditProductName(productID, product)
	if err != nil {
		log.Println("error EditProducNametHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (api ProductAPI) DeleteProductByIDHandler(context *gin.Context) {
	productID := context.Param("product_id")
	err := api.ProductRepository.DeleteProductByID(productID)
	if err != nil {
		log.Println("error DeleteProductHandler", err.Error())
		context.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
	context.JSON(http.StatusNoContent, gin.H{"message": "success"})
}
