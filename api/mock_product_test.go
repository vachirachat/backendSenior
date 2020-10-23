package api_test

import (
	"backendSenior/model"
	"time"

	"github.com/globalsign/mgo/bson"
)

type mockProductRepository struct{}

const (
	listProduct = `{"products":[{"product_id":"5beaf7bd62e63844ce22cc58","product_name":"CocaCola","product_price":"14.00","amount":20,"updated_time":"0001-01-01T00:00:00Z"},{"product_id":"5beaf7bd62e63844ce22cc57","product_name":"M150","product_price":"10.00","amount":50,"updated_time":"0001-01-01T00:00:00Z"}]}`
)

func (productRepository mockProductRepository) GetAllProduct() ([]model.Product, error) {
	updatedTime, _ := time.Parse("2006-01-02", "20060102")
	return []model.Product{
		{
			ProductID:    bson.ObjectIdHex("5beaf7bd62e63844ce22cc58"),
			ProductName:  "CocaCola",
			ProductPrice: "14.00",
			Amount:       20,
			UpdatedTime:  updatedTime,
		},
		{
			ProductID:    bson.ObjectIdHex("5beaf7bd62e63844ce22cc57"),
			ProductName:  "M150",
			ProductPrice: "10.00",
			Amount:       50,
		},
	}, nil
}

func (productRepository mockProductRepository) AddProduct(product model.Product) error {
	return nil
}

func (productRepository mockProductRepository) EditProductName(productID string, product model.Product) error {
	return nil
}
func (productRepository mockProductRepository) GetProductByID(productID string) (model.Product, error) {
	return model.Product{
		ProductID:    bson.ObjectIdHex("5befe40d9c71fe169a4341df"),
		ProductName:  "M150",
		ProductPrice: "14.00",
		Amount:       20,
	}, nil
}

func (productRepository mockProductRepository) GetLastProduct() (model.Product, error) {
	return model.Product{}, nil
}

func (productRepository mockProductRepository) DeleteProductByID(productID string) error {
	return nil
}
