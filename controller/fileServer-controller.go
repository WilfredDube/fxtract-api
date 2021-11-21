package controller

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"

// 	// "github.com/WilfredDube/fxtract-backend/lib/helper"
// 	"github.com/dgrijalva/jwt-go"
// 	"github.com/gorilla/mux"
// )

// type fileServer struct{ jwtService: service.JWTService}

// type FileServer interface {
// 	Download(w http.ResponseWriter, r *http.Request)
// }

// func YourHandler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Gorilla!\n")
// }

// func UnauthorizedHandler(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(401)
// 	fmt.Fprintf(w, "401 Unauthorized\n")
// }

// func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(404)
// 	fmt.Fprintf(w, "404 Not Found\n")
// }

// func (f *fileServer) Download(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	token, err := f.jwtService.GetAuthenticationToken(r, "fxtract")
// 	if err != nil {
// 		response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
// 		w.WriteHeader(http.StatusForbidden)
// 		json.NewEncoder(w).Encode(response)
// 		return
// 	}

// 	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 		mux := mux.NewRouter()
// 		mux.HandleFunc("/", YourHandler)
// 		d := "/static/"
// 		mux.HandleFunc(d, UnauthorizedHandler)                           // prevent directory listing
// 		mux.HandleFunc(d+"{filename:[a-zA-Z0-9]+}.gif", NotFoundHandler) // black list file type
// 		mux.PathPrefix(d).Handler(http.StripPrefix(d, http.FileServer(http.Dir("."+d))))
// 	}
// }
