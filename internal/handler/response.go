package handler

import (
	"net/http"
)

const (
	errorCodeBadRequest    = "bad_request"
	errorCodeValidation    = "validation_error"
	errorCodeInternalError = "internal_error"
	errorCodeNotFound      = "not_found"
	errorCodeUnauthorized  = "unauthorized"
	errorCodeForbidden     = "forbidden"
	errorCodeConflict      = "conflict"
)

type envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *errorBody `json:"error,omitempty"`
	Meta    *meta      `json:"meta,omitempty"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

type errorDetail struct {
	Field   string `json:"field,omitempty"`
	Rule    string `json:"rule,omitempty"`
	Message string `json:"message"`
}

type meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

func ok(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, envelope{Success: true, Data: data})
}

func created(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, envelope{Success: true, Data: data})
}

func okPaginated(w http.ResponseWriter, data any, page, limit, total int) {
	totalPages := 0
	if limit > 0 {
		totalPages = total / limit
		if total%limit != 0 {
			totalPages++
		}
	}

	writeJSON(w, http.StatusOK, envelope{
		Success: true,
		Data:    data,
		Meta: &meta{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	})
}

func badRequest(w http.ResponseWriter, code, msg string, details any) {
	writeJSON(w, http.StatusBadRequest, envelope{
		Success: false,
		Error: &errorBody{
			Code:    code,
			Message: msg,
			Details: details,
		},
	})
}

func unprocessableEntity(w http.ResponseWriter, details any) {
	writeJSON(w, http.StatusUnprocessableEntity, envelope{
		Success: false,
		Error: &errorBody{
			Code:    errorCodeValidation,
			Message: "request validation failed",
			Details: details,
		},
	})
}

func internalError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusInternalServerError, envelope{
		Success: false,
		Error: &errorBody{
			Code:    errorCodeInternalError,
			Message: msg,
		},
	})
}

func notFound(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusNotFound, envelope{
		Success: false,
		Error: &errorBody{
			Code:    errorCodeNotFound,
			Message: msg,
		},
	})
}

func unauthorized(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusUnauthorized, envelope{
		Success: false,
		Error: &errorBody{
			Code:    errorCodeUnauthorized,
			Message: msg,
		},
	})
}

func forbidden(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusForbidden, envelope{
		Success: false,
		Error: &errorBody{
			Code:    errorCodeForbidden,
			Message: msg,
		},
	})
}

func conflict(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusConflict, envelope{
		Success: false,
		Error: &errorBody{
			Code:    errorCodeConflict,
			Message: msg,
		},
	})
}
