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
	Config *config.Config
}

func NewApi(conf *config.Config) *Api {
	a := &Api{
		Router: mux.NewRouter(),
		Config: conf,
	}

	a.Router.HandleFunc("/", index).Methods("GET")
	a.Router.HandleFunc("/v1/celestial/{regionid}/{clusterid}/nodes", a.clusterNodes).Methods("GET")

	return a
}

func (self *Api) newClient() *client.Client {
	return client.NewClient(self.Config.GetClientConfig())
}

func (self *Api) formatPath(params ...string) string {
	return ""
}

func (self *Api) Run() {
	log.Fatal(http.ListenAndServe(self.Config.GetApiAddress(), self.Router))
}

func (self *Api) clusterNodes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	regionid := vars["regionid"]
	clusterid := vars["clusterid"]

	client := self.newClient()
	defer client.Close()

	var nodes []string
	for n := range client.PrintClusterNodes(regionid, clusterid) {
		nodes = append(nodes, n)
	}

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
