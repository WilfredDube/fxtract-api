package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/lib/helper"
	persistence "github.com/WilfredDube/fxtract-backend/repository/reposelect"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

const (
	ALLUSERS = "allusers"
)

type userController struct {
	userService service.UserService
	jwtService  service.JWTService
	cache       *redis.Client
}

// UserController -
type UserController interface {
	Update(w http.ResponseWriter, r *http.Request)
	Promote(w http.ResponseWriter, r *http.Request)
	Profile(w http.ResponseWriter, r *http.Request)
	GetAllUsers(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

// NewUserController -
func NewUserController(service service.UserService, jwtService service.JWTService, cache *redis.Client) UserController {
	return &userController{
		userService: service,
		jwtService:  jwtService,
		cache:       cache,
	}
}

// Update - add a new user
func (c *userController) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	user := &entity.User{}
	err = json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["user_id"].(string)
		uid, _ := primitive.ObjectIDFromHex(userID)

		user.ID = uid
		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		user.Password = string(hash)

		n, err := strconv.ParseInt(claims["created_at"].(string), 10, 64)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		user.CreatedAt = n

		u, err := c.userService.Update(user)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		userData := NewLoginResponse(u)

		go persistence.ClearCache(userID)

		response := helper.BuildResponse(true, "OK", userData)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := helper.BuildErrorResponse("Failed to process request", "User not found", helper.EmptyObj{})
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(response)
}

func (c *userController) Promote(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		params := mux.Vars(r)
		op := r.FormValue("operation")

		uid := params["id"]

		user, err := c.userService.Profile(uid)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if op == "promote" {
			user.UserRole = entity.ADMIN
		} else if op == "demote" {
			user.UserRole = entity.GENERAL_USER
		}

		u, err := c.userService.Update(user)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to promote user", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		userData := NewLoginResponse(u)

		go persistence.ClearCache(ALLUSERS)

		response := helper.BuildResponse(true, "OK", userData)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := helper.BuildErrorResponse("Failed to process request", "User not found", helper.EmptyObj{})
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(response)
}

// Profile -
func (c *userController) Profile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised user", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		id := claims["user_id"].(string)

		result, err := c.cache.Get(id).Result()

		var user *entity.User
		if err != nil {
			user, err = c.userService.Profile(id)
			if err != nil {
				response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response)
				return
			}

			bytes, err := json.Marshal(user)
			if err != nil {
				response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}

			if err := c.cache.Set(id, bytes, 30*time.Minute).Err(); err != nil {
				response := helper.BuildErrorResponse("Failed to cache request", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}
		} else {
			json.Unmarshal([]byte(result), &user)
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

func (c *userController) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		_ = claims["user_id"].(string)

		cachedUsers, err := c.cache.Get(ALLUSERS).Result()

		var users []entity.User
		if err != nil {
			users, err = c.userService.GetAll()
			if err != nil {
				response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(response)
				return
			}

			bytes, err := json.Marshal(users)
			if err != nil {
				response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}

			if err := c.cache.Set(ALLUSERS, bytes, 30*time.Second).Err(); err != nil {
				response := helper.BuildErrorResponse("Failed to cache request", err.Error(), helper.EmptyObj{})
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}
		} else {
			json.Unmarshal([]byte(cachedUsers), &users)
		}

		var searchedUsers []entity.User

		if query := r.FormValue("q"); query != "" {
			query = strings.ToLower(query)
			for _, user := range users {
				if strings.Contains(strings.ToLower(user.Firstname), query) || strings.Contains(strings.ToLower(user.Lastname), query) {
					searchedUsers = append(searchedUsers, user)
				}
			}
		} else {
			searchedUsers = users
		}

		res := helper.BuildResponse(true, "OK", searchedUsers)
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

	token, err := c.jwtService.GetAuthenticationToken(r, "fxtract")
	if err != nil {
		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["user_id"].(string)
		params := mux.Vars(r)
		id := params["id"]

		deleteCount, err := c.userService.Delete(id)
		if err != nil {
			response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		if deleteCount == 0 {
			response := helper.BuildResponse(true, "User not found", helper.EmptyObj{})
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		go persistence.ClearCache(userID)
		go persistence.ClearCache(ALLUSERS)

		response := helper.BuildResponse(true, "User deletion successful", helper.EmptyObj{})
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}

	response := helper.BuildErrorResponse("Failed to process request", "User not found", helper.EmptyObj{})
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(response)
}
