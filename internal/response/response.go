package response

import (
	"encoding/json"
	"net/http"
)

const (
	CodeSuccess       = "00"
	CodeBadRequest    = "01"
	CodeNotFound      = "02"
	CodeInternalError = "03"
)

type Envelope struct {
	ResponseCode string `json:"responseCode"`
	ResponseDesc string `json:"responseDesc"`
	ResponseData any    `json:"responseData,omitempty"`
}

func Success(w http.ResponseWriter, httpStatus int, desc string, data any) {
	write(w, httpStatus, Envelope{
		ResponseCode: CodeSuccess,
		ResponseDesc: desc,
		ResponseData: data,
	})
}

func Error(w http.ResponseWriter, httpStatus int, code string, desc string) {
	write(w, httpStatus, Envelope{
		ResponseCode: code,
		ResponseDesc: desc,
	})
}

func write(w http.ResponseWriter, httpStatus int, envelope Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(envelope)
}
