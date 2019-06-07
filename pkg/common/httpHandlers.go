package common

import (
	"encoding/json"
	"log"
	"net/http"
)

///////////////////////////////////////////////////////////
// WS handler
func WSHome(w http.ResponseWriter, r *http.Request) {
	HandleStatusNotFoundResponse(w, "")
}

///////////////////////////////////////////////////////////
// handler
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	json.NewEncoder(w).Encode("Still alive!")
}

func HandleMethodNotAllowedResponse(w http.ResponseWriter, customError string) {
	log.Printf("Error 405: %v\n", customError)
	// send a bad request error back to the caller
	var msg string
	if msg = customError; msg == "" {
		msg = "Method Not Allowed"
	}
	errorStatus := http.StatusMethodNotAllowed
	response := HTTPResponseWrapper{
		Ok:      false,
		Status:  errorStatus,
		Message: msg,
	}
	HandleHTTPResponse(w, response)
}

func HandleForbiddenResponse(w http.ResponseWriter, customError string) {
	log.Printf("Error 403: %v\n", customError)
	// send a bad request error back to the caller
	var msg string
	if msg = customError; msg == "" {
		msg = "Forbidden"
	}
	errorStatus := http.StatusForbidden
	response := HTTPResponseWrapper{
		Ok:      false,
		Status:  errorStatus,
		Message: msg,
	}
	HandleHTTPResponse(w, response)
}

func HandleStatusNotFoundResponse(w http.ResponseWriter, customError string) {
	log.Printf("Error 404: %s\n", customError)
	// send a not found error back to the caller
	var msg string
	if msg = customError; msg == "" {
		msg = "Page Not Found"
	}
	errorStatus := http.StatusNotFound
	response := HTTPResponseWrapper{
		Ok:      false,
		Status:  errorStatus,
		Message: msg,
	}
	HandleHTTPResponse(w, response)
}

func HandleServerErrorResponse(w http.ResponseWriter, err error) {
	log.Printf("Error 500: %v\n", err)
	// send a internal server error back to the caller
	errorStatus := http.StatusInternalServerError
	response := HTTPResponseWrapper{
		Ok:      false,
		Status:  errorStatus,
		Message: http.StatusText(http.StatusInternalServerError),
		Error:   err,
	}
	HandleHTTPResponse(w, response)
}

func HandleStatus400Response(w http.ResponseWriter, customError string) {
	log.Printf("Error 400: %v\n", customError)
	// send a bad request error back to the caller
	var msg string
	if msg = customError; msg == "" {
		msg = "Bad Request"
	}
	errorStatus := http.StatusBadRequest
	response := HTTPResponseWrapper{
		Ok:      false,
		Status:  errorStatus,
		Message: msg,
	}
	HandleHTTPResponse(w, response)
}

func HandleStatusOKResponse(w http.ResponseWriter, data HTTPResult) {
	log.Printf("Response 200: %v\n", data)

	httpStatus := http.StatusOK
	response := HTTPResponseWrapper{
		Ok:      true,
		Status:  httpStatus,
		Message: http.StatusText(http.StatusOK),
		Result:  data,
	}
	HandleHTTPResponse(w, response)
}

func HandleHTTPResponse(w http.ResponseWriter, data HTTPResponseWrapper) {
	w.WriteHeader(data.Status)
	// write json data
	json.NewEncoder(w).Encode(data)
}
