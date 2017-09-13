package contract

//This file is auto-generated by go-raml
//Do not edit this file by hand since it will be overwritten during the next generation

import (
	"github.com/gorilla/mux"
	"net/http"
)

type UsersusernamecontractsInterface interface {
	// Get the contracts where the user is 1 of the parties. Order descending by date.
	// It is handler for GET /users/{username}/contracts
	Get(http.ResponseWriter, *http.Request)
}

func UsersusernamecontractsInterfaceRoutes(r *mux.Router, i UsersusernamecontractsInterface) {
	r.HandleFunc("/users/{username}/contracts", i.Get).Methods("GET")
}