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
	"github.com/WilfredDube/fxtract-backend/lib/helper"
	msgqueue_amqp "github.com/WilfredDube/fxtract-backend/lib/msgqueue/amqp"
	"github.com/WilfredDube/fxtract-backend/listener"
	"github.com/WilfredDube/fxtract-backend/middleware"
	"github.com/WilfredDube/fxtract-backend/repository"
	persistence "github.com/WilfredDube/fxtract-backend/repository/reposelect"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/streadway/amqp"
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

	conn, err := amqp.Dial(config.AMQPMessageBroker)
	if err != nil {
		panic(err.Error())
	}

	eventEmitter, err := msgqueue_amqp.NewAMQPEventEmitter(conn, "processes")
	if err != nil {
		panic(err)
	}

	eventListener, err := msgqueue_amqp.NewAMQPEventListener(conn, "processes", "FEATURERECOGNITIONCOMPLETE")
	if err != nil {
		panic(err)
	}

	processPlannerEventListener, err := msgqueue_amqp.NewAMQPEventListener(conn, "processes", "PROCESSPLANNINGCOMPLETE")
	if err != nil {
		panic(err)
	}

	cache := persistence.SetUpRedis()

	JWTService := service.NewJWTService(store)

	userRepo := repository.NewUserRepository(*repo)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService, JWTService, cache)

	authService := service.NewAuthService(userRepo)
	verificationRepo := repository.NewVerificationRepository(*repo)
	verificationService := service.NewVerificationService(verificationRepo)

	mailService := service.NewSGMailService(&config)
	authController := controller.NewAuthController(authService, JWTService, mailService, verificationService)

	cadFileRepo := repository.NewCadFileRepository(*repo)
	cadFileService := service.NewCadFileService(cadFileRepo)

	processingPlanRepo := repository.NewProcessingPlanRepository(*repo)
	processingPlanService := service.NewProcessingPlanService(processingPlanRepo)

	projectRepo := repository.NewProjectRepository(*repo)
	projectService := service.NewProjectService(projectRepo)
	projectController := controller.NewProjectController(projectService, userService, cadFileService, processingPlanService, JWTService, cache)

	cadFileController := controller.NewCADFileController(cadFileService, projectService, JWTService)

	toolRepo := repository.NewToolRepository(*repo)
	toolService := service.NewToolService(toolRepo)
	toolController := controller.NewToolController(toolService, userService, JWTService, cache)

	materialRepo := repository.NewMaterialRepository(*repo)
	materialService := service.NewMaterialService(materialRepo)
	materialController := controller.NewMaterialController(materialService, userService, JWTService, cache)

	taskRepo := repository.NewTaskRepository(*repo)
	taskService := service.NewTaskService(taskRepo)
	taskController := controller.NewTaskController(taskService, JWTService, cache)

	freController := controller.NewFREController(config, cadFileService, processingPlanService, userService, JWTService, taskService, cache, eventEmitter)

	r := mux.NewRouter()

	// Project creation and CAD file upload
	r.HandleFunc("/api/user/projects", projectController.AddProject).Methods("POST")
	r.HandleFunc("/api/user/projects", projectController.UpdateProject).Methods("PUT")
	r.HandleFunc("/api/user/projects", projectController.FindAll).Methods("GET")
	r.HandleFunc("/api/user/projects/{id}", projectController.FindByID).Methods("GET")
	r.HandleFunc("/api/user/projects/{id}", projectController.Delete).Methods("DELETE")
	r.HandleFunc("/api/user/projects/{id}", projectController.Upload).Methods("POST").Queries("operation", "{upload}")
	r.HandleFunc("/api/user/projects/{id}/files", projectController.FindAllCADFiles).Methods("GET")
	r.HandleFunc("/api/user/projects/{pid}/files/{id}", projectController.FindCADFileByID).Methods("GET")
	r.HandleFunc("/api/user/projects/{pid}/files/{id}", projectController.DeleteCADFile).Methods("DELETE")

	// Feature recognition / processing plan API based on the CAD file's process level
	r.HandleFunc("/api/user/projects/{pid}/files/{id}", freController.ProcessCADFile).Methods("POST").Queries("operation", "{process}")
	r.HandleFunc("/api/user/projects/{pid}/files", freController.BatchProcessCADFiles).Methods("POST").Queries("operation", "{process}")
	r.HandleFunc("/api/user/process/{id}", projectController.FindProcessPlan).Methods("GET")

	// User registration and login
	r.HandleFunc("/api/auth/register", authController.Register).Methods("POST")
	r.HandleFunc("/api/auth/verify", authController.VerifyMail).Methods("POST")
	r.HandleFunc("/api/auth/get-password-reset-code", authController.GeneratePassResetCode).Methods("POST")
	r.HandleFunc("/api/auth/verify-password-reset-code", authController.VerifyPasswordReset).Methods("POST")
	r.HandleFunc("/api/auth/reset-password", authController.ResetPassword).Methods("POST")
	r.HandleFunc("/api/auth/verify", authController.VerifyMail).Methods("POST")
	r.HandleFunc("/api/auth/login", authController.Login).Methods("POST")
	r.HandleFunc("/api/auth/logout", authController.Logout).Methods("POST")

	// User account update and profile
	r.HandleFunc("/api/user", userController.Update).Methods("PUT")
	r.HandleFunc("/api/user/profile", userController.Profile).Methods("GET")

	/* ---------------- Admin endpoints ------------------*/
	r.HandleFunc("/api/admin/users", middleware.CheckAdminRole(JWTService, userController.GetAllUsers)).Methods("GET")
	r.HandleFunc("/api/admin/users", middleware.CheckAdminRole(JWTService, authController.Register)).Methods("POST")
	r.HandleFunc("/api/admin/users/{id}", middleware.CheckAdminRole(JWTService, userController.Promote)).Methods("PUT")

	// Tool creation
	r.HandleFunc("/api/admin/tools", middleware.CheckAdminRole(JWTService, toolController.AddTool)).Methods("POST")
	r.HandleFunc("/api/admin/tools", middleware.CheckAdminRole(JWTService, toolController.FindAll)).Methods("GET")
	r.HandleFunc("/api/admin/tools/{id}", middleware.CheckAdminRole(JWTService, toolController.FindByID)).Methods("GET")
	r.HandleFunc("/api/admin/tools/angle/{angle}", middleware.CheckAdminRole(JWTService, toolController.FindByAngle)).Methods("GET")
	r.HandleFunc("/api/admin/tools/{id}", middleware.CheckAdminRole(JWTService, toolController.Delete)).Methods("DELETE")

	// Material creation
	r.HandleFunc("/api/admin/materials", middleware.CheckAdminRole(JWTService, materialController.AddMaterial)).Methods("POST")
	r.HandleFunc("/api/admin/materials", middleware.CheckAdminRole(JWTService, materialController.FindAll)).Methods("GET")
	r.HandleFunc("/api/admin/materials/{id}", middleware.CheckAdminRole(JWTService, materialController.Find)).Methods("GET")
	r.HandleFunc("/api/admin/materials/{id}", middleware.CheckAdminRole(JWTService, materialController.Delete)).Methods("DELETE")

	// Files uploaded
	r.HandleFunc("/api/admin/files", middleware.CheckAdminRole(JWTService, cadFileController.FindAllFiles)).Methods("GET")

	// Tasks
	r.HandleFunc("/api/admin/task", middleware.CheckAdminRole(JWTService, taskController.FindAll)).Methods("GET")
	r.HandleFunc("/api/admin/task/{id}", middleware.CheckAdminRole(JWTService, taskController.Find)).Methods("GET")

	// processes: type, status

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

	processor := listener.EventProcessor{EventListener: eventListener, CadFileService: cadFileService,
		TaskService: taskService, ProcessingPlanService: processingPlanService, ToolService: toolService, MaterialService: materialService}
	go processor.ProcessEvents("featureRecognitionComplete")

	processPlanner := listener.EventProcessor{EventListener: processPlannerEventListener, CadFileService: cadFileService,
		TaskService: taskService, ProcessingPlanService: processingPlanService, MaterialService: materialService}
	go processPlanner.ProcessEvents("processPlanningComplete")

	fmt.Printf("Terminated %s\n", <-errs)
}
