package common

import (
	"net/http"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

func NewRouter() []Route {

	commonRouter := []Route{
		Route{"HealthCheck", "GET", "/healthcheck", HealthCheck},
	}

	return commonRouter
	//router.HandleFunc("/healthcheck", healthCheck).Methods("GET")
}

func NewWSRouter() []Route {

	commonWSRouter := []Route{
		Route{"Home", "GET", "/", WSHome},
	}

	return commonWSRouter
	//router.HandleFunc("/healthcheck", healthCheck).Methods("GET")
}
