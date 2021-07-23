package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/GrooveCommunity/go-dispatcher/entity"

	"os"

	"github.com/gorilla/mux"

	"github.com/GrooveCommunity/go-dispatcher/internal"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", handleHealthy).Methods("GET")
	router.HandleFunc("/dispatcher/healthy", handleValidateHealthy).Methods("GET")
	router.HandleFunc("/dispatcher/put-rule", handlePutRule).Methods("POST")

	log.Println("Port: ", os.Getenv("APP_PORT"))

	go internal.ForwardIssue(os.Getenv("JIRA_USERNAME"), os.Getenv("JIRA_TOKENAPI"), os.Getenv("JIRA_ENDPOINT"))

	log.Fatal(http.ListenAndServe(":"+os.Getenv("APP_PORT"), router))
}

func handleHealthy(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(entity.Healthy{Status: "Deu certo!"})
}

func handleValidateHealthy(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(entity.Healthy{Status: "Success!"})
}

func handlePutRule(w http.ResponseWriter, r *http.Request) {
	var rule entity.Rule

	err := json.NewDecoder(r.Body).Decode(&rule)
	if err != nil {
		panic(err)
	}

	internal.WriteRule(rule)
}
