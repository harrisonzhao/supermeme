package main

import (
	"github.com/gorilla/mux"
	"github.com/harrisonzhao/supermeme/controllers"
	"github.com/harrisonzhao/supermeme/shared/constants"
	"github.com/harrisonzhao/supermeme/shared/db"
	"net/http"
)

func main() {
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
	//http.ListenAndServe(":3000", nil)
	http.ListenAndServeTLS(":443", "keys/www.catchupbot.com/fullchain.pem", "keys/www.catchupbot.com/privkey.pem", nil)
}
