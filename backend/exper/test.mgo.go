package main

import (
	"fmt"
)

func printError(err error) {
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}
	fmt.Println("OK")
}

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

func main() {
	var a = []int{0, 1, 2, 3, 4, 5}
	a[0], a[4] = a[4], a[0]
	fmt.Printf("%#v\n", a)
	a[4], a[4] = a[4], a[4]
	fmt.Printf("%#v\n", a)
}
