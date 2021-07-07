package controller

import (
	"encoding/json"
	"net/http"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/helper"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userController struct {
	userService service.UserService
	jwtService  service.JWTService
}

// UserController -
type UserController interface {
	Update(w http.ResponseWriter, r *http.Request)
	Profile(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// NewUserController -
func NewUserController(service service.UserService, jwtService service.JWTService) UserController {
	return &userController{
		userService: service,
		jwtService:  jwtService,
	}
}

// Update - add a new user
func (c *userController) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user := &entity.User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	authHeader := r.Header.Get("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)
	if errToken != nil {
		panic(errToken.Error())
	}

	claims := token.Claims.(jwt.MapClaims)
	uid, err := primitive.ObjectIDFromHex(claims["id"].(string))
	user.ID = uid

	u, err := c.userService.Update(user)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := helper.BuildResponse(true, "OK!", u)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Profile -
func (c *userController) Profile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authHeader := r.Header.Get("Authorization")
	token, errToken := c.jwtService.ValidateToken(authHeader)
	if errToken != nil {
		response := helper.BuildErrorResponse("Failed to process request", errToken.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id := claims["user_id"].(string)

		user, err := c.userService.Profile(id)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		userData := NewLoginResponse(user)

		res := helper.BuildResponse(true, "OK", userData)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
		return
	}

	response := helper.BuildErrorResponse("Failed to process request", "User not found", helper.EmptyObj{})
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(response)
}

// Delete -
func (c *userController) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id := params["id"]

	deleteCount, err := c.userService.Delete(id)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if 0 == deleteCount {
		response := helper.BuildResponse(true, "User not found", helper.EmptyObj{})
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := helper.BuildResponse(true, "User deletion successful", helper.EmptyObj{})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
