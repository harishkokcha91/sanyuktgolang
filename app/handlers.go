package app

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
)

type Customer struct {
	Name string `json:"full_name" xml:"name"`
	City string `json:"city_name" xml:"city_name"`
}

func greet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello world")
}

func getAllCustomer(w http.ResponseWriter, r *http.Request) {
	customer := []Customer{
		{"Harish", "Delhi"},
		{"Harish k", "New Delhi"},
	}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(customer)
}

func getAllCustomerXml(w http.ResponseWriter, r *http.Request) {
	customer := []Customer{
		{"Harish", "Delhi"},
		{"Harish k", "New Delhi"},
	}
	if r.Header.Get("Content-Type") == "application/xml" {

		w.Header().Add("Content-Type", "application/xml")
		xml.NewEncoder(w).Encode(customer)
	} else {
		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(customer)
	}
}
