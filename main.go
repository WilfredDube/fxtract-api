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
	"github.com/streadway/amqp"
	"gopkg.in/boj/redistore.v1"
)

var (
	configPath = flag.String("c", "./configuration/config.json", "Set the configuration file for setting up the database.")
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

	sessionStore, err := redistore.NewRediStore(0, "tcp", config.RedisHost+":"+config.RedisPort, "", []byte("oNrT10hnwnnUTeiwqm1ISP6W5qXmHWkT"))
	if err != nil {
		panic(err)
	}

	JWTService := service.NewJWTService(sessionStore)

	redisCache := persistence.SetUpRedis(config)

	userRepo := repository.NewUserRepository(*repo)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService, JWTService, redisCache)

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
	projectController := controller.NewProjectController(projectService, userService, cadFileService, processingPlanService, JWTService, redisCache)

	cadFileController := controller.NewCADFileController(cadFileService, projectService, JWTService, processingPlanService, redisCache)

	toolRepo := repository.NewToolRepository(*repo)
	toolService := service.NewToolService(toolRepo)
	toolController := controller.NewToolController(toolService, userService, JWTService, redisCache)

	materialRepo := repository.NewMaterialRepository(*repo)
	materialService := service.NewMaterialService(materialRepo)
	materialController := controller.NewMaterialController(materialService, userService, JWTService, redisCache)

	taskRepo := repository.NewTaskRepository(*repo)
	taskService := service.NewTaskService(taskRepo)
	taskController := controller.NewTaskController(taskService, JWTService, redisCache)

	processorController := service.NewProcessor(JWTService)
	freController := controller.NewFREController(config, cadFileService, processingPlanService, userService, JWTService, taskService, redisCache, eventEmitter, processorController)

	r := mux.NewRouter()

	d := "/objs/"
	r.PathPrefix(d).Handler(http.StripPrefix(d, http.FileServer(http.Dir("."+d))))

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
	r.HandleFunc("/api/user/ws", processorController.Handler(freController.BatchProcessCADFiles)).Methods("GET") //.Methods("POST").Queries("operation", "{process}")
	// r.HandleFunc("/api/user/ws", freController.BatchProcessCADFiles).Methods("GET") //.Methods("POST").Queries("operation", "{process}")
	r.HandleFunc("/api/user/process/{id}", projectController.FindProcessPlan).Methods("GET")

	r.HandleFunc("/api/user/materials", materialController.FindAll).Methods("GET")

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
	r.HandleFunc("/api/admin/materials/{id}", middleware.CheckAdminRole(JWTService, materialController.Find)).Methods("GET")
	r.HandleFunc("/api/admin/materials/{id}", middleware.CheckAdminRole(JWTService, materialController.Delete)).Methods("DELETE")

	// Files uploaded
	r.HandleFunc("/api/admin/files", middleware.CheckAdminRole(JWTService, cadFileController.FindAllFiles)).Methods("GET")

	// Download OBJ file
	r.HandleFunc("/api/user/projects/files", cadFileController.DownloadOBJ).Methods("POST").Queries("url", "{url}")

	// Tasks
	r.HandleFunc("/api/admin/tasks", middleware.CheckAdminRole(JWTService, taskController.FindAll)).Methods("GET")
	r.HandleFunc("/api/user/tasks", taskController.FindByUserID).Methods("GET")
	r.HandleFunc("/api/tasks/{id}", taskController.Find).Methods("GET")

	// processes: type, status

	originsObj := handlers.AllowedOrigins([]string{"*"})
	credentials := handlers.AllowCredentials()
	headersObj := handlers.AllowedHeaders([]string{"Origin", "Access-Control, Allow-Origin", "Content-Type",
		"Accept", "Authorization", "Origin, Accept", "X-Requested-With",
		"Access-Control-Request-Method", "Access-Control-Request-Header",
	})
	methodsObj := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	server := handlers.CORS(credentials, originsObj, methodsObj, headersObj)(r)

	processorController.Start()

	errs := make(chan error, 3)
	go func() {
		fmt.Printf("Listening on port: %v\n", config.RestfulEndPoint)
		errs <- http.ListenAndServe(config.RestfulEndPoint, server)
	}()

	// go func() {
	// 	fmt.Printf("Listening on port: %v\n", config.RestfulTLSEndPoint)
	// 	errs <- http.ListenAndServeTLS(config.RestfulTLSEndPoint, "cert/cert.pem", "cert/key.pem", server)
	// }()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	processor := listener.EventProcessor{EventListener: eventListener, CadFileService: cadFileService,
		TaskService: taskService, ProcessingPlanService: processingPlanService, ToolService: toolService, MaterialService: materialService, ProjectService: projectService, UserService: userService, Processor: processorController}
	go processor.ProcessEvents("featureRecognitionComplete")

	processPlanner := listener.EventProcessor{EventListener: processPlannerEventListener, CadFileService: cadFileService,
		TaskService: taskService, ProcessingPlanService: processingPlanService, MaterialService: materialService, ProjectService: projectService, UserService: userService, Processor: processorController}
	go processPlanner.ProcessEvents("processPlanningComplete")

	fmt.Printf("Terminated %s\n", <-errs)
}
