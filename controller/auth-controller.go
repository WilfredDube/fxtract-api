package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/lib/helper"
	"github.com/WilfredDube/fxtract-backend/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type loginResponse struct {
	Firstname string      `json:"firstname"`
	Lastname  string      `json:"lastname"`
	Email     string      `json:"email"`
	UserRole  entity.Role `json:"role"`
	CreatedAt int64       `json:"created_at"`
}

// AuthController -
type AuthController interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
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
		UserRole:  user.UserRole,
		CreatedAt: user.CreatedAt,
	}
}

func isEmailValid(email string) bool {
	if len(email) < 5 && len(email) > 254 {
		return false
	}

	var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !emailRegex.MatchString(email) {
		return false
	}

	parts := strings.Split(email, "@")
	mx, err := net.LookupMX(parts[1])
	if err != nil || len(mx) == 0 {
		return false
	}
	return true
}

func isPasswordValid(s string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(s) >= 7 {
		hasMinLen = true
	}
	for _, char := range s {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

func validate(user *entity.User) error {
	if (user.Email == "") || (user.Password == "") {
		return errors.New("email or password can not be empty")
	}

	if !isEmailValid(user.Email) {
		return errors.New("invalid email address")
	}

	if !isPasswordValid(user.Password) {
		return errors.New("invalid password")
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
	if dup {
		response := helper.BuildErrorResponse("Failed to process request", "Duplicate email", helper.EmptyObj{})
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(response)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		response := helper.BuildErrorResponse("Failed user registration", "User registration failed", helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	user.ID = primitive.NewObjectID()
	user.Password = string(hash)
	user.CreatedAt = time.Now().Unix()
	if user.UserRole == 0 {
		user.UserRole = entity.GENERAL_USER
	}

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

	err = c.jwtService.SetAuthentication(&authResult, "fxtract", 86400*7, service.LOGIN, w, r)
	if err != nil {
		response := helper.BuildErrorResponse("Please check again your credential", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	userData := NewLoginResponse(&authResult)
	response := helper.BuildResponse(true, "OK!", userData)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (c *authController) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := c.jwtService.SetAuthentication(&entity.User{}, "fxtract", -1, service.LOGOUT, w, r)
	if err != nil {
		response := helper.BuildErrorResponse("Already logged off", "Invalid procedure", helper.EmptyObj{})
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := helper.BuildResponse(true, "OK!", "Logout successful")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
