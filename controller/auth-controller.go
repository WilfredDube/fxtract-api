package controller

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/lib/helper"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/mazen160/go-random"
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

type changePasswordMessage struct {
	Email string `json:"email"`
}
type verificationMessage struct {
	Email string `json:"email"`
	Code  string `json:"code" `
}

type passwordRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
	Code            string `json:"code" `
}

// AuthController -
type AuthController interface {
	Register(w http.ResponseWriter, r *http.Request)
	VerifyMail(w http.ResponseWriter, r *http.Request)
	GeneratePassResetCode(w http.ResponseWriter, r *http.Request)
	VerifyPasswordReset(w http.ResponseWriter, r *http.Request)
	ResetPassword(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type authController struct {
	authService  service.AuthService
	jwtService   service.JWTService
	mailService  service.MailService
	verification service.VerificationService
}

//NewAuthController creates a new instance of AuthController
func NewAuthController(authService service.AuthService, jwtService service.JWTService,
	mailService service.MailService, verification service.VerificationService) AuthController {
	return &authController{
		authService:  authService,
		jwtService:   jwtService,
		mailService:  mailService,
		verification: verification,
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

// GenerateRandomString generate a string of random characters of given length
func GenerateRandomString(length int) string {
	charset := random.ASCIICharacters
	code, err := random.Random(length, charset, true)
	if err != nil {
		log.Fatal(err)
	}
	return code
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

	from := os.Getenv("SENDER_EMAIL_ADDRESS")
	if from == "" {
		response := helper.BuildErrorResponse("Failed to process request. Please try again later.", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	_, err = c.authService.CreateUser(*user)
	if err != nil {
		response := helper.BuildErrorResponse("Failed user registration", "User registration failed", helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Send verification mail
	to := []string{user.Email}
	subject := "Email Verification for Fxtract"
	mailType := service.MailConfirmation
	mailData := &service.MailData{
		Username: user.Firstname,
		Code:     GenerateRandomString(8),
	}

	mailReq := c.mailService.NewMail(from, to, subject, mailType, mailData)
	err = c.mailService.SendMail(mailReq)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	verificationData := &entity.Verification{
		Email:     user.Email,
		Code:      mailData.Code,
		Type:      entity.MailConfirmation,
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(45)).Unix(),
	}

	_, err = c.verification.Create(verificationData)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := helper.BuildResponse(true, "User created successfully", helper.EmptyObj{})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (c *authController) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user := &entity.User{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", "", helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	authResult, err := c.authService.VerifyCredential(user.Email, user.Password)
	if err != nil {
		response := helper.BuildErrorResponse("Wrong email or password", "", helper.EmptyObj{})
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	if !authResult.IsVerified {
		response := helper.BuildErrorResponse("Please check email for the verification code", "Account has not been verified", helper.EmptyObj{})
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	err = c.jwtService.SetAuthentication(&authResult, "fxtract", 86400*7, service.LOGIN, w, r)
	if err != nil {
		response := helper.BuildErrorResponse("Wrong email or password", "", helper.EmptyObj{})
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

func (c *authController) VerifyMail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	verificationMsg := &verificationMessage{}

	err := json.NewDecoder(r.Body).Decode(verificationMsg)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	verificationData := &entity.Verification{
		Email: verificationMsg.Email,
		Code:  verificationMsg.Code,
		Type:  entity.MailConfirmation,
	}

	actualVerificationData, err := c.verification.Find(verificationData.Email, verificationData.Type)
	if err != nil {
		response := helper.BuildErrorResponse("unable to fetch verification data", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(response)
		return
	}

	valid, err := c.verify(actualVerificationData, verificationData)
	if !valid {
		response := helper.BuildErrorResponse("Invalid verification code", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(response)
		return
	}

	// delete the VerificationData from db
	_, err = c.verification.Delete(actualVerificationData.ID.Hex())
	if err != nil {
		response := helper.BuildErrorResponse("Unable to verify mail. Please try again later", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
	}

	// correct code, update user status to verified.
	err = c.authService.UpdateUserVerificationStatus(verificationData.Email, true)
	if err != nil {
		response := helper.BuildErrorResponse("Unable to verify mail. Please try again later", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := helper.BuildResponse(true, "OK!", "Mail Verification succeeded")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (c *authController) verify(actualVerificationData *entity.Verification, verificationData *entity.Verification) (bool, error) {
	timeT := time.Unix(actualVerificationData.ExpiresAt, 0)

	// check for expiration
	if timeT.Before(time.Now()) {
		log.Println()
		_, err := c.verification.Delete(actualVerificationData.ID.Hex())
		if err != nil {
			return false, errors.New("verification data provided is expired")
		}

		return false, errors.New("confirmation code has expired. Please try generating a new code")
	}

	if actualVerificationData.Code != verificationData.Code {
		log.Println("verification of mail failed. Invalid verification code provided")
		return false, errors.New("verification code provided is Invalid. Please look in your mail for the code")
	}

	return true, nil
}

func (c *authController) GeneratePassResetCode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	passwordMsg := &changePasswordMessage{}
	err := json.NewDecoder(r.Body).Decode(passwordMsg)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	user := c.authService.FindByEmail(passwordMsg.Email)
	if err != nil {
		response := helper.BuildErrorResponse("User not found", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Send verification mail
	from := os.Getenv("SENDER_EMAIL_ADDRESS")
	if from == "" {
		response := helper.BuildErrorResponse("Failed to process request. Please try again later.", from, helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	to := []string{user.Email}
	subject := "Email Verification for Fxtract"
	mailType := service.PassReset
	mailData := &service.MailData{
		Username: user.Firstname,
		Code:     GenerateRandomString(8),
	}

	mailReq := c.mailService.NewMail(from, to, subject, mailType, mailData)
	err = c.mailService.SendMail(mailReq)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	verificationData := &entity.Verification{
		Email:     user.Email,
		Code:      mailData.Code,
		Type:      entity.PassReset,
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(45)).Unix(),
	}

	_, err = c.verification.Create(verificationData)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := helper.BuildResponse(true, "Check your email for the code to reset your password", helper.EmptyObj{})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (c *authController) VerifyPasswordReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	verificationMsg := &verificationMessage{}

	err := json.NewDecoder(r.Body).Decode(verificationMsg)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	verificationData := &entity.Verification{
		Email: verificationMsg.Email,
		Code:  verificationMsg.Code,
		Type:  entity.PassReset,
	}

	actualVerificationData, err := c.verification.Find(verificationData.Email, verificationData.Type)
	if err != nil {
		response := helper.BuildErrorResponse("unable to fetch verification data", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(response)
		return
	}

	valid, err := c.verify(actualVerificationData, verificationData)
	if !valid {
		response := helper.BuildErrorResponse("Invalid verification code", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := helper.BuildResponse(true, "OK!", actualVerificationData.Code)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (c *authController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	passwdReq := &passwordRequest{}

	err := json.NewDecoder(r.Body).Decode(passwdReq)
	if err != nil {
		response := helper.BuildErrorResponse("Failed to process request", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	verificationData, err := c.verification.Find(passwdReq.Email, entity.PassReset)
	if err != nil {
		response := helper.BuildErrorResponse("unable to fetch verification data", err.Error(), helper.EmptyObj{})
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(response)
		return
	}

	if verificationData.Code != passwdReq.Code {
		response := helper.BuildErrorResponse("Unable to reset password. Please try again later", "Invalid reset code", helper.EmptyObj{})
		log.Println("verification code did not match even after verifying PassReset")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	if passwdReq.Password != passwdReq.PasswordConfirm {
		response := helper.BuildErrorResponse("Password and re-entered Password are not same", "password and password re-enter did not match", helper.EmptyObj{})
		log.Println("password and password re-enter did not match")
		w.WriteHeader(http.StatusNotAcceptable)
		json.NewEncoder(w).Encode(response)
		return
	}

	if !isPasswordValid(passwdReq.Password) {
		response := helper.BuildErrorResponse("Invalid password. Please try again later", "Password does not meet criteria", helper.EmptyObj{})
		log.Println("Password does not meet criteria")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(passwdReq.Password), bcrypt.DefaultCost)
	if err != nil {
		response := helper.BuildErrorResponse("Failed password change", "Failed to change the password. Try again later", helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	hashedPassword := string(hash)

	// delete the VerificationData from db
	_, err = c.verification.Delete(verificationData.ID.Hex())
	if err != nil {
		response := helper.BuildErrorResponse("Failed", "Failed to change the password. Try again later", helper.EmptyObj{})
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	err = c.authService.UpdateUserPassword(passwdReq.Email, hashedPassword)
	if err != nil {
		response := helper.BuildErrorResponse("Unable to verify mail. Please try again later", err.Error(), helper.EmptyObj{})
		log.Println("unable to set user verification status to true")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := helper.BuildResponse(true, "OK!", "Password change successfully")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
