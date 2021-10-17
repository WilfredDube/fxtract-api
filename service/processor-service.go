package service

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/lib/helper"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
)

type Processor struct {
	Users            map[string]*websocket.Conn
	ProcessorChannel chan string
	TaskChannel      chan *entity.Task
	CADFilesChannel  chan CADFileResponse
	jwtService       JWTService
}

type CADFileResponse struct {
	UserID   string
	CadFiles []entity.CADFile
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
)

func NewProcessor(jwtService JWTService) *Processor {
	return &Processor{
		jwtService:       jwtService,
		Users:            make(map[string]*websocket.Conn),
		ProcessorChannel: make(chan string),
		TaskChannel:      make(chan *entity.Task),
		CADFilesChannel:  make(chan CADFileResponse),
	}
}

func (p *Processor) Handler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		token, err := p.jwtService.GetAuthenticationToken(r, "fxtract")
		if err != nil {
			response := helper.BuildErrorResponse("Unauthorised", "User not authenticated", helper.EmptyObj{})
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			id := claims["user_id"].(string)

			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Fatal("connection err:", err)
			}

			// if _, ok := p.Users[id]; !ok {
			p.Users[id] = conn
			log.Printf("Add user: %v Total: %d\n", id, len(p.Users))
			// }
		}

		next(w, r)
	}
}

func (p *Processor) Run() {
	for {
		select {
		case cadFiles := <-p.ProcessorChannel:
			go func(cadFiles string) {
				log.Println(cadFiles)
			}(cadFiles)

		case task := <-p.TaskChannel:
			if conn, ok := p.Users[task.UserID.Hex()]; ok {
				response := helper.BuildResponse(true, "File processing complete", task)
				resp, err := json.Marshal(response)
				if err != nil {
					log.Println(err.Error())
				}
				conn.WriteMessage(websocket.TextMessage, resp)
				log.Printf("Send task: %v user: %v Total: %d\n", task.ID.Hex(), task.UserID.Hex(), len(p.Users))
			}
		case cadFilesResponse := <-p.CADFilesChannel:
			if conn, ok := p.Users[cadFilesResponse.UserID]; ok {
				response := helper.BuildResponse(true, "files", cadFilesResponse.CadFiles)
				resp, err := json.Marshal(response)
				if err != nil {
					log.Println(err.Error())
				}
				conn.WriteMessage(websocket.TextMessage, resp)
				log.Printf("Send cad files; user: %v Total: %d\n", cadFilesResponse.UserID, len(cadFilesResponse.CadFiles))
			}
		}
	}
}

func (p *Processor) Start() {
	go p.Run()
}
