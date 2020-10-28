package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

type ProductInfo struct {
	Product []Product `json:"products"`
}

type Product struct {
	ProductID    bson.ObjectId `json:"productId" bson:"_id,omitempty"`
	ProductName  string        `json:"productName" bson:"productName"`
	ProductPrice string        `json:"productPrice" bson:"productPrice"`
	Amount       int           `json:"amount" bson:"amount"`
	CreatedTime  time.Time     `json:"-" bson:"createdTime"`
	UpdatedTime  time.Time     `json:"updatedTime" bson:"updatedTime"`
}
