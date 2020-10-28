package repository

import (
	"backendSenior/model"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
)

type ProductRepository interface {
	GetAllProduct() ([]model.Product, error)
	GetLastProduct() (model.Product, error)
	GetProductByID(productID string) (model.Product, error)
	AddProduct(product model.Product) error
	EditProductName(productID string, product model.Product) error
	DeleteProductByID(productID string) error
}

type ProductRepositoryMongo struct {
	ConnectionDB *mgo.Session
}

const (
	DBName     = "myShop"
	collection = "product"
)

func (productMongo ProductRepositoryMongo) GetAllProduct() ([]model.Product, error) {
	var products []model.Product
	err := productMongo.ConnectionDB.DB(DBName).C(collection).Find(nil).All(&products)
	return products, err
}

func (productMogo ProductRepositoryMongo) GetProductByID(productID string) (model.Product, error) {
	var product model.Product
	objectID := bson.ObjectIdHex(productID)
	err := productMogo.ConnectionDB.DB(DBName).C(collection).FindId(objectID).One(&product)
	return product, err
}
func (productMogo ProductRepositoryMongo) GetLastProduct() (model.Product, error) {
	var product model.Product
	err := productMogo.ConnectionDB.DB(DBName).C(collection).Find(nil).Sort("-created_time").One(&product)
	return product, err
}
func (productMongo ProductRepositoryMongo) AddProduct(product model.Product) error {
	return productMongo.ConnectionDB.DB(DBName).C(collection).Insert(product)
}

func (productMongo ProductRepositoryMongo) EditProductName(productID string, product model.Product) error {
	objectID := bson.ObjectIdHex(productID)
	newName := bson.M{"$set": bson.M{"product_name": product.ProductName, "updated_time": time.Now()}}
	return productMongo.ConnectionDB.DB(DBName).C(collection).UpdateId(objectID, newName)
}

func (productMongo ProductRepositoryMongo) DeleteProductByID(productID string) error {
	objectID := bson.ObjectIdHex(productID)
	return productMongo.ConnectionDB.DB(DBName).C(collection).RemoveId(objectID)
}
