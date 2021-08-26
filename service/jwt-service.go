package service

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
)

// DBTYPE -
type AUTHTYPE string

// Const -
const (
	LOGIN  AUTHTYPE = "login"
	LOGOUT AUTHTYPE = "logout"
)

//JWTService is a contract of what jwtService can do
type JWTService interface {
	GenerateToken(userID string) string
	ValidateToken(token string) (*jwt.Token, error)
	SetAuthentication(userID string, cookieName string, maxAge int, authType AUTHTYPE, w http.ResponseWriter, r *http.Request) error
	GetAuthenticationToken(r *http.Request, cookieName string) *jwt.Token
}

type jwtCustomClaim struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

type jwtService struct {
	secretKey string
	issuer    string
}

//NewJWTService method is creates a new instance of JWTService
func NewJWTService() JWTService {
	return &jwtService{
		issuer:    "Fxtract",
		secretKey: getSecretKey(),
	}
}

func getSecretKey() string {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey != "" {
		secretKey = "Fxtract"
	}
	return secretKey
}

func (j *jwtService) GenerateToken(UserID string) string {
	claims := &jwtCustomClaim{
		UserID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().AddDate(1, 0, 0).Unix(),
			Issuer:    j.issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		panic(err)
	}
	return t
}

func (j *jwtService) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t_ *jwt.Token) (interface{}, error) {
		if _, ok := t_.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method %v", t_.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})
}

func (j *jwtService) SetAuthentication(useid string, cookieName string, maxAge int, authType AUTHTYPE, w http.ResponseWriter, r *http.Request) error {
	session, err := j.store.Get(r, cookieName)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	session.Options = &sessions.Options{
		HttpOnly: true,
		// Secure:   true,
		MaxAge: maxAge,
		Path:   "/",
	}

	if authType == LOGIN {
		generatedToken := j.GenerateToken(useid)
		session.Values["authenticated"] = true
		session.Values["token"] = generatedToken
	} else {
		if session.Values["authenticated"] == false {
			return fmt.Errorf("already signed out")
		}
		session.Values["authenticated"] = false
		session.Values["token"] = ""
	}

	err = sessions.Save(r, w)
	if err != nil {
		return err
	}

	return nil
}

func (j *jwtService) GetAuthenticationToken(r *http.Request, cookieName string) *jwt.Token {
	session, err := j.store.Get(r, cookieName)
	if err != nil {
		return nil
	}

	// Check if user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		fmt.Print(ok)
		return nil
	}

	authHeader := session.Values["token"].(string)
	token, errToken := j.ValidateToken(authHeader)
	if errToken != nil {
		return nil
	}

	return token
}
