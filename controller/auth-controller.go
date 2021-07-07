package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/helper"
	"github.com/WilfredDube/fxtract-backend/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type loginResponse struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	CreatedAt int64  `json:"created_at"`
}

// AuthController -
type AuthController interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
}

type authController struct {
	authService service.AuthService
	jwtService  service.JWTService
}

//NewAuthController creates a new instance of AuthController
func NewAuthController(authService service.AuthService, jwtService service.JWTService) AuthController {
	return &authController{
		authService: authService,
		jwtService:  jwtService,
	}
}

// NewLoginResponse return user details without sensitive password data
func NewLoginResponse(user *entity.User) loginResponse {
	return loginResponse{
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
		Email:     user.Email,
		Token:     user.Token,
		CreatedAt: user.CreatedAt,
	}
}

func validate(user *entity.User) error {
	if (user.Email == "") || (user.Password == "") {
		return errors.New("email or password can not be empty")
	}

	return nil
}

func (c *authController) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user := &entity.User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	err = validate(user)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	dup := c.authService.IsDuplicateEmail(user.Email)
	if dup == true {
		response := helper.BuildErrorResponse("Failed to process request", "Duplicate email", helper.EmptyObj{})
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.ID = primitive.NewObjectID()
	user.Password = string(hash)
	user.CreatedAt = time.Now().Unix()
	user.Token = c.jwtService.GenerateToken(user.ID.Hex())

	_, err = c.authService.CreateUser(*user)
	if err != nil {
		response := helper.BuildErrorResponse("Failed user registration", "User registration failed", helper.EmptyObj{})
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	helper.CreateFolder(user.ID.Hex(), false)

	response := helper.BuildResponse(true, "OK!", helper.EmptyObj{})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (c *authController) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user := &entity.User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	authResult, err := c.authService.VerifyCredential(user.Email, user.Password)
	if err != nil {
		response := helper.BuildErrorResponse("Please check again your credential", "Invalid Credential", helper.EmptyObj{})
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	generatedToken := c.jwtService.GenerateToken(authResult.ID.Hex())
	authResult.Token = generatedToken

	userData := NewLoginResponse(&authResult)
	response := helper.BuildResponse(true, "OK!", userData)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
