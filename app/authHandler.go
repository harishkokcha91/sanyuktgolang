package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sanyuktgolang/logger"
	"sanyuktgolang/model"
	"sanyuktgolang/service"
)

type AuthHandler struct {
	service service.AuthService
}

func (h AuthHandler) NotImplementedHandler(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, http.StatusOK, "Handler not implemented...")
}

func (h AuthHandler) GenerateOtp(w http.ResponseWriter, r *http.Request) {
	var loginRequest model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		logger.Error("Error while decoding login request: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)

	} else {

		response, appErr := h.service.GenerateOtp(loginRequest)
		if appErr != nil {
			writeResponse(w, appErr.Code, appErr.AsMessage().Message)
		} else {
			writeResponse(w, http.StatusOK, *response)
		}

	}
}

func (h AuthHandler) VerifyOtp(w http.ResponseWriter, r *http.Request) {
	var loginRequest model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		logger.Error("Error while decoding login request: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
	} else {
		response, appErr := h.service.VerifyOtp(loginRequest)
		if appErr != nil {
			writeResponse(w, appErr.Code, appErr.AsMessage().Message)
		} else {
			writeResponse(w, http.StatusOK, *response)
		}

	}
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		logger.Error("Error while decoding login request: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
	} else {
		token, appErr := h.service.Login(loginRequest)
		if appErr != nil {
			writeResponse(w, appErr.Code, appErr.AsMessage().Message)
		} else {
			writeResponse(w, http.StatusOK, *token)
		}
	}
}

func (h AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var loginRequest model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		logger.Error("Error while decoding login request: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
	} else {
		token, appErr := h.service.UpdateUser(loginRequest)
		if appErr != nil {
			writeResponse(w, appErr.Code, appErr.AsMessage().Message)
		} else {
			writeResponse(w, http.StatusOK, *token)
		}
	}
}

/*
	Sample URL string

http://localhost:8181/auth/verify?token=somevalimodelkenstring&routeName=GetCustomer&customer_id=2000&account_id=95470
*/
func (h AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	urlParams := make(map[string]string)

	// converting from Query to map type
	for k := range r.URL.Query() {
		urlParams[k] = r.URL.Query().Get(k)
	}

	if urlParams["token"] != "" {
		appErr := h.service.Verify(urlParams)
		if appErr != nil {
			writeResponse(w, appErr.Code, notAuthorizedResponse(appErr.Message))
		} else {
			writeResponse(w, http.StatusOK, authorizedResponse())
		}
	} else {
		writeResponse(w, http.StatusForbidden, notAuthorizedResponse("missing token"))
	}
}

func (h AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var refreshRequest model.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&refreshRequest); err != nil {
		logger.Error("Error while decoding refresh token request: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
	} else {
		token, appErr := h.service.Refresh(refreshRequest)
		if appErr != nil {
			writeResponse(w, appErr.Code, appErr.AsMessage().Message)
		} else {
			writeResponse(w, http.StatusOK, *token)
		}
	}
}

func notAuthorizedResponse(msg string) map[string]interface{} {
	return map[string]interface{}{
		"isAuthorized": false,
		"message":      msg,
	}
}

func authorizedResponse() map[string]bool {
	return map[string]bool{"isAuthorized": true}
}

func writeResponse(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	print(code)
	print(fmt.Sprint(data))
	if code != http.StatusOK {
		if str, ok := data.(string); ok {
			data = model.Response{Status: false, StatusCode: uint16(code), Message: string(str)}
		}
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		panic(err)
	}
}
