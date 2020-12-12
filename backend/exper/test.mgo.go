package main

import (
	"encoding/json"
	"fmt"

	"github.com/globalsign/mgo/bson"
)

// func printError(err error) {
// 	if err != nil {
// 		fmt.Println("Error: ", err.Error())
// 		return
// 	}
// 	fmt.Println("OK")
// }

// func main() {
// 	mongo, err := mgo.Dial("mongodb://127.0.0.1")
// 	if err != nil {
// 		log.Fatalf("Error connecting, %v", err.Error())
// 	}

// 	col := mongo.DB("foo").C("room")

// 	id := bson.NewObjectId()
// 	id2 := bson.NewObjectId()
// 	err = col.Insert(bson.M{
// 		"_id":  id,
// 		"data": "foo",
// 	})
// 	printError(err)

// 	var res interface{}
// 	err = col.FindId(id2).One(&res)
// 	printError(err)

// 	err = col.UpdateId(id2, bson.M{
// 		"$set": bson.M{
// 			"foo": "bar",
// 		},
// 	})
// 	printError(err)

// 	err = col.RemoveId(id2)
// 	printError(err)

// 	id3 := bson.NewObjectId()
// 	id4 := bson.NewObjectId()
// 	err = col.Insert(bson.M{
// 		"_id":  id3,
// 		"data": "foo",
// 	}, bson.M{
// 		"_id":  id4,
// 		"data": "foo",
// 	})
// 	printError(err)

// 	info, err := col.UpdateAll(bson.M{
// 		"_id": bson.M{
// 			"$in": []bson.ObjectId{id3, id4},
// 		},
// 	}, bson.M{
// 		"$set": bson.M{
// 			"data": "bar",
// 		},
// 	})
// 	printError(err)
// 	fmt.Printf("update 2 result %#v\n", info)

// 	id5 := bson.NewObjectId()
// 	info, err = col.UpdateAll(bson.M{
// 		"_id": bson.M{
// 			"$in": []bson.ObjectId{id5, id4},
// 		},
// 	}, bson.M{
// 		"$set": bson.M{
// 			"data": "bar",
// 		},
// 	})
// 	printError(err)
// 	fmt.Printf("update 2 result with non exists %#v\n", info)

// 	col.Find(nil).Count()

// }

type testUnmarshal struct {
	OID bson.ObjectId `json:"oid"`
}

func main() {
	jsonStr := []byte(`{"oid": "5fd342cee2f8760aceb47a64"}`)
	var test testUnmarshal
	err := json.Unmarshal(jsonStr, &test)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}
	fmt.Printf("%+v\n", test)
	fmt.Println(test.OID.Hex())
	jsonStr, err = json.Marshal(test)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}
	fmt.Printf("%s\n", jsonStr)
}
