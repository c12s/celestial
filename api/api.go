package api

import (
	"encoding/json"
	"fmt"
	"github.com/c12s/celestial/client"
	"github.com/c12s/celestial/config"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Api struct {
	Router *mux.Router
	Client *client.Client
	Conf   *config.ConnectionConfig
}

func NewApi(conf *config.Config) *Api {
	a := &Api{
		Router: mux.NewRouter(),
		Client: client.NewClient(conf.GetClientConfig()),
		Conf:   conf.GetConnectionConfig(),
	}

	a.Router.HandleFunc("/", index).Methods("GET")
	a.Router.HandleFunc("/v1/celestial/{regionid}/{clusterid}/nodes", a.clusterNodes).Methods("GET")

	return a
}

func (self *Api) formatPath(params ...string) string {

	return ""
}

func (self *Api) Close() {
	self.Client.Close()
}

func (self *Api) Run() {
	defer self.Close()
	log.Fatal(http.ListenAndServe(self.Conf.GetApiAddress(), self.Router))
}

func (self *Api) clusterNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	vars := mux.Vars(r)
	regionid := vars["regionid"]
	clusterid := vars["clusterid"]

	var nodes []string
	for n := range self.Client.PrintClusterNodes(regionid, clusterid) {
		nodes = append(nodes, n)
	}

	// fmt.Fprintln(w, "Hello")

	sendJSONResponse(w, nodes)
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
