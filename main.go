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
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	configPath = flag.String("c", "./configuration/config.json", "Set the configuration file for setting up the database.")
	store      = sessions.NewCookieStore([]byte("oNrT10hnwnnUTeiwqm1ISP6W5qXmHWkT"))
)

func main() {
	flag.Parse()

	helper.CreateFolder("", true)

	config, _ := configuration.ExtractConfiguration(*configPath)
	repo := persistence.NewPersistenceLayer(config)

	JWTService := service.NewJWTService(store)

	userRepo := repository.NewUserRepository(*repo)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService, JWTService)

	authService := service.NewAuthService(userRepo)
	authController := controller.NewAuthController(authService, JWTService)

	cadFileRepo := repository.NewCadFileRepository(*repo)
	cadFileService := service.NewCadFileService(cadFileRepo)

	processingPlanRepo := repository.NewProcessingPlanRepository(*repo)
	processingPlanService := service.NewProcessingPlanService(processingPlanRepo)

	projectRepo := repository.NewProjectRepository(*repo)
	projectService := service.NewProjectService(projectRepo)
	projectController := controller.NewProjectController(projectService, userService, cadFileService, processingPlanService, JWTService)

	// cadFileController := controller.NewCADFileController(cadFileService, projectService, JWTService)

	toolRepo := repository.NewToolRepository(*repo)
	toolService := service.NewToolService(toolRepo)
	toolController := controller.NewToolController(toolService, userService, JWTService)

	materialRepo := repository.NewMaterialRepository(*repo)
	materialService := service.NewMaterialService(materialRepo)
	materialController := controller.NewMaterialController(materialService, userService, JWTService)

	freController := controller.NewFREController(config, cadFileService, processingPlanService, userService, JWTService)

	r := mux.NewRouter()

	// Project creation and CAD file upload
	r.HandleFunc("/projects", projectController.AddProject).Methods("POST")
	r.HandleFunc("/projects", projectController.FindAll).Methods("GET")
	r.HandleFunc("/projects/{id}", projectController.FindByID).Methods("GET")
	r.HandleFunc("/projects/{id}", projectController.Delete).Methods("DELETE")
	r.HandleFunc("/projects/{id}", projectController.Upload).Methods("POST").Queries("operation", "{upload}")
	r.HandleFunc("/projects/{id}/files", projectController.FindAllCADFiles).Methods("GET")
	r.HandleFunc("/projects/{pid}/files/{id}", projectController.FindCADFileByID).Methods("GET")
	r.HandleFunc("/projects/{pid}/files/{id}", projectController.DeleteCADFile).Methods("DELETE")

	// Feature recognition / processing plan API based on the CAD file's process level
	r.HandleFunc("/projects/{pid}/files/{id}", freController.ProcessCADFile).Methods("POST").Queries("operation", "{process}")
	r.HandleFunc("/projects/{pid}/files", freController.BatchProcessCADFiles).Methods("POST").Queries("operation", "{process}")

	// Tool creation
	r.HandleFunc("/tools", toolController.AddTool).Methods("POST")
	r.HandleFunc("/tools", toolController.FindAll).Methods("GET")
	r.HandleFunc("/tools/{id}", toolController.FindByID).Methods("GET")
	r.HandleFunc("/tools/{id}", toolController.FindByAngle).Methods("GET")
	r.HandleFunc("/tools/{id}", toolController.Delete).Methods("DELETE")

	// Material creation
	r.HandleFunc("/materials", materialController.AddMaterial).Methods("POST")
	r.HandleFunc("/materials", materialController.FindAll).Methods("GET")
	r.HandleFunc("/materials/{id}", materialController.Find).Methods("GET")
	r.HandleFunc("/materials/{id}", materialController.Delete).Methods("DELETE")

	// User registration and login
	r.HandleFunc("/api/auth/register", authController.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", authController.Login).Methods("POST")
	r.HandleFunc("/api/auth/logout", authController.Logout).Methods("POST")

	// User account update and profile
	r.HandleFunc("/api/user", userController.Update).Methods("POST")
	r.HandleFunc("/api/user/profile", userController.Profile).Methods("GET")

	originsObj := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	headersObj := handlers.AllowedHeaders([]string{"Origin", "Access-Control, Allow-Origin", "Content-Type", "Accept", "Authorization", "Origin, Accept", "X-Requested-With", "Access-Control-Request-Method", "Access-Control-Request-Header"})
	methodsObj := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	server := handlers.CORS(originsObj, methodsObj, headersObj)(r)

	errs := make(chan error, 3)
	go func() {
		fmt.Printf("Listening on port: %v\n", config.RestfulEndPoint)
		errs <- http.ListenAndServe(config.RestfulEndPoint, server)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	notificationService := notification.NewMQResponseController(config, cadFileService, processingPlanService, userService, JWTService)
	go func() {
		notificationService.FeatureRecognitionNotifications("FRECRESPONSE")
	}()

	go func() {
		notificationService.ProcessingPlanNotifications("PPRESPONSE")
	}()

	fmt.Printf("Terminated %s\n", <-errs)
}
