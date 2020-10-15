package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection
var ctx = context.TODO()

func StartDBMongo() {
	// clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	// client, err := mongo.Connect(ctx, clientOptions)

	// err = client.Ping(ctx, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// collection = client.Database("tasker").Collection("tasks")
	// log.Println("already connect to Mongo server")

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017/"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from the api!")
}

type Trainer struct {
	Name string `json:”name,omitempty”`
	Age  int32  `json:”age,omitempty”`
	City string `json:”city,omitempty”`
}

// func InsertPost(title string, body string) {
// 	post := Post{title, body}
// 	collection := client.Database("my_database").Collection("posts")
// 	insertResult, err := collection.InsertOne(context.TODO(), post)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println("Inserted post with ID:", insertResult.InsertedID)
// }

// func GetPost(id bson.ObjectId) {
// 	collection := client.Database("my_database").Collection("posts")
// 	filter := bson.D
// 	var post Post
// 	err := collection.FindOne(context.TODO(), filter).Decode(&post)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println("Found post with title ", post.Title)
// }

func main() {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	collection := client.Database("test").Collection("trainers")
	fmt.Println(collection)

	// ash := Trainer{"Ash", 10, "Pallet Town"}
	misty := Trainer{"Misty", 10, "Cerulean City"}
	brock := Trainer{"Brock", 15, "Pewter City"}

	// insertResult, err := collection.InsertOne(context.TODO(), ash)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	insertResult2, err := collection.InsertOne(context.TODO(), misty)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult2.InsertedID)

	insertResult3, err := collection.InsertOne(context.TODO(), brock)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult3.InsertedID)

	filter := bson.D{{"name", "Ash"}}
	var result Trainer

	err = collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found a single document: %+v\n", result)

	// insertData("google", 20, "usa")

	filter2 := bson.D{{"name", "Misty"}}
	var result2 Trainer

	err = collection.FindOne(context.TODO(), filter2).Decode(&result2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found a single document: %+v\n", result2)

}

func insertData(name string, age int32, city string) {
	ash := Trainer{"sdasd", 10, "Pallet Town"}
	insertResult, err := collection.InsertOne(context.TODO(), &ash)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
}
