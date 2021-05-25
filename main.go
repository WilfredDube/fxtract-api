package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/WilfredDube/fxtract-backend/configuration"
	"github.com/WilfredDube/fxtract-backend/controller"
	"github.com/WilfredDube/fxtract-backend/helper"
	"github.com/WilfredDube/fxtract-backend/repository"
	persistence "github.com/WilfredDube/fxtract-backend/repository/reposelect"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/gorilla/mux"
)

var (
	configPath = flag.String("c", "./configuration/config.json", "Set the configuration file for setting up the database.")
)

func main() {
	flag.Parse()

	helper.CreateFolder("", true)

	config, _ := configuration.ExtractConfiguration(*configPath)
	repo := persistence.NewPersistenceLayer(config)

	JWTService := service.NewJWTService()

	projectRepo := repository.NewProjectRepository(*repo)
	projectService := service.NewProjectService(projectRepo)
	projectController := controller.NewProjectController(projectService, JWTService)

	userRepo := repository.NewUserRepository(*repo)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService, JWTService)

	r := mux.NewRouter()

	r.HandleFunc("/projects", projectController.AddProject).Methods("POST")
	r.HandleFunc("/projects", projectController.FindAll).Methods("GET")
	r.HandleFunc("/projects/{id}", projectController.FindByID).Methods("GET")
	r.HandleFunc("/projects/{id}", projectController.Delete).Methods("DELETE")

	authService := service.NewAuthService(userRepo)
	authController := controller.NewAuthController(authService, JWTService)
	r.HandleFunc("/api/auth/register", authController.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", authController.Login).Methods("POST")

	r.HandleFunc("/api/user/profile", userController.Profile).Methods("GET")
	// r.HandleFunc("/register", userController.AddUser).Methods("POST")
	// r.HandleFunc("/register", userController.AddUser).Methods("POST")

	errs := make(chan error, 3)
	go func() {
		fmt.Printf("Listening on port: %v", config.RestfulEndPoint)
		errs <- http.ListenAndServe(config.RestfulEndPoint, r)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	fmt.Printf("Terminated %s\n", <-errs)
}
