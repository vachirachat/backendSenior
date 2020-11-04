package main

import (
	"backendSenior/route"
	"backendSenior/utills"
	"log"

	"github.com/globalsign/mgo"
)

func main() {
	connectionDB, err := mgo.Dial(utills.MONGOENDPOINT)
	if err != nil {
		log.Panic("Can no connect Database", err.Error())
	}
	router := route.NewRouter(connectionDB)
	router.Run(utills.PORTWEBSERVER)

}

// func serveDefault(w http.ResponseWriter, r *http.Request) {
// 	log.Println(r.URL)
// 	if r.URL.Path != "/" {
// 		http.Error(w, "Not found", http.StatusNotFound)
// 		return
// 	}
// 	if r.Method != "GET" {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	http.ServeFile(w, r, "index.html")
// }

// func main() {
// 	hub := H
// 	go hub.Run()
// 	http.HandleFunc("/", serveDefault)
// 	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
// 		ServeWs(w, r)
// 	})
// 	//Listerning on port :8080...
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }
