package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type App struct {
	Router  *mux.Router
	Address string
}

func NewApi(address string) *App {
	a := App{
		Router:  mux.NewRouter(),
		Address: address,
	}

	a.Router.HandleFunc("/hello/", index).Methods("GET")

	return &a
}

func (self *App) Run() {
	log.Fatal(http.ListenAndServe(self.Address, self.Router))
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, "Hello")
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to encode a JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		log.Printf("Failed to write the response body: %v", err)
		return
	}
}
