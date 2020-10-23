package repository_test

import (
	"backendSenior/model"
	"backendSenior/repository"
	"log"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/globalsign/mgo"
	"github.com/stretchr/testify/assert"
)

const mogoDBEnPint = "mongodb://localhost:27017"

func connectionDB() *mgo.Session {
	connectionDB, err := mgo.Dial(mogoDBEnPint)
	if err != nil {
		log.Panic("Con not connect to database", err.Error())
	}
	return connectionDB
}

func Test_AddProduct_Shold_Be_Product(t *testing.T) {
	connectionDB := connectionDB()
	defer connectionDB.Close()
	productRepository := repository.ProductRepositoryMongo{
		ConnectionDB: connectionDB,
	}
	fixedTime, _ := time.Parse("2006-Jan-02", "2018-Oct-08")
	product := model.Product{
		ProductID:    bson.ObjectIdHex("5befe40d9c71fe169a4341df"),
		ProductName:  "M150",
		ProductPrice: "14.00",
		Amount:       20,
		CreatedTime:  fixedTime,
		UpdatedTime:  fixedTime,
	}

	actual := productRepository.AddProduct(product)

	if actual != nil {
		t.Error("Con not Add product", actual.Error())
	}
}

func Test_GetAllProduct_Should_Be_Array_Product(t *testing.T) {
	connectionDB := connectionDB()
	defer connectionDB.Close()
	fixedTime, _ := time.Parse("2006-Jan-02", "2018-Oct-08")
	expected := []model.Product{
		{
			ProductID:    bson.ObjectIdHex("5befe40d9c71fe169a4341df"),
			ProductName:  "M150",
			ProductPrice: "14.00",
			Amount:       20,
			CreatedTime:  fixedTime,
			UpdatedTime:  fixedTime,
		},
	}
	productRepository := repository.ProductRepositoryMongo{
		ConnectionDB: connectionDB,
	}

	actual, _ := productRepository.GetAllProduct()

	assert.Equal(t, expected, actual)
}

func Test_EditProductName_Input_Product_Name_M150_Should_Be_Edited(t *testing.T) {
	connectionDB := connectionDB()
	defer connectionDB.Close()
	productID := "5befe40d9c71fe169a4341df"
	product := model.Product{
		ProductName: "M150",
	}
	productRepository := repository.ProductRepositoryMongo{
		ConnectionDB: connectionDB,
	}

	actual := productRepository.EditProductName(productID, product)

	if actual != nil {
		t.Error("Con not Edit product", actual.Error())
	}
}

func Test_GetProductByID_Input_ID_5befe40d9c71fe169a4341df_Should_Be_Product_Name_M150(t *testing.T) {
	connectionDB := connectionDB()
	defer connectionDB.Close()
	fixedTime, _ := time.Parse("2006-Jan-02", "2018-Oct-08")
	expected := model.Product{
		ProductID:    bson.ObjectIdHex("5befe40d9c71fe169a4341df"),
		ProductName:  "M150",
		ProductPrice: "14.00",
		Amount:       20,
		CreatedTime:  fixedTime,
		UpdatedTime:  fixedTime,
	}
	productID := "5befe40d9c71fe169a4341df"
	productRepository := repository.ProductRepositoryMongo{
		ConnectionDB: connectionDB,
	}

	actual, _ := productRepository.GetProductByID(productID)

	assert.Equal(t, expected, actual)
}

func Test_GetLastProduct_Should_Be_Be_Product_Name_M150(t *testing.T) {
	connectionDB := connectionDB()
	defer connectionDB.Close()
	fixedTime, _ := time.Parse("2006-Jan-02", "2018-Oct-08")
	expected := model.Product{
		ProductID:    bson.ObjectIdHex("5befe40d9c71fe169a4341df"),
		ProductName:  "M150",
		ProductPrice: "14.00",
		Amount:       20,
		CreatedTime:  fixedTime,
		UpdatedTime:  fixedTime,
	}
	productRepository := repository.ProductRepositoryMongo{
		ConnectionDB: connectionDB,
	}

	actual, _ := productRepository.GetLastProduct()

	assert.Equal(t, expected, actual)
}
func Test_DeleteProductByID_Input_Id_5befe40d9c71fe169a4341df_Should_Be_Deleted(t *testing.T) {
	connectionDB := connectionDB()
	defer connectionDB.Close()
	productID := "5befe40d9c71fe169a4341df"
	productRepository := repository.ProductRepositoryMongo{
		ConnectionDB: connectionDB,
	}

	actua := productRepository.DeleteProductByID(productID)

	if actua != nil {
		t.Error("Con not delete product", actua.Error())
	}
}
