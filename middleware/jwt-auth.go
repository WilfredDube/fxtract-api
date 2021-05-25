package middleware

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/WilfredDube/fxtract-backend/helper"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/dgrijalva/jwt-go"
)

//AuthorizeJWT validates the token user given, return 401 if not valid
func AuthorizeJWT(jwtService service.JWTService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response := helper.BuildErrorResponse("Failed to process request", "No token found", nil)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}
		token, err := jwtService.ValidateToken(authHeader)
		if token.Valid {
			claims := token.Claims.(jwt.MapClaims)
			log.Println("Claim[user_id]: ", claims["user_id"])
			log.Println("Claim[issuer] :", claims["issuer"])
		} else {
			log.Println(err)
			response := helper.BuildErrorResponse("Token is not valid", err.Error(), nil)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(response)
		}
	}
}
