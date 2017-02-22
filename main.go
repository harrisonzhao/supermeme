package main

import (
	"flag"
	"github.com/gorilla/mux"
	"github.com/harrisonzhao/supermeme/controllers"
	"github.com/harrisonzhao/supermeme/shared/constants"
	"github.com/harrisonzhao/supermeme/shared/db"
	"log"
	"net/http"
)

func main() {
	flag.Parse()
	dbutil.InitDb(constants.DbName)
	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/webhook", controllers.InitMessenger().Handler).Methods("POST")
	r.HandleFunc("/webhook", controllers.MessengerWebhook).Methods("GET")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, constants.PublicDir+"/index.html")
	})
	fs := http.FileServer(http.Dir(constants.PublicDir))
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", fs))
	http.Handle("/", r)
	//log.Fatal(http.ListenAndServe(":3000", nil))
	log.Fatal(http.ListenAndServeTLS(":443", "keys/answerme.me/fullchain.pem", "keys/answerme.me/privkey.pem", nil))
}
