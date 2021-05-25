FROM golang:latest as build

RUN apk add --no-cache git

RUN mkdir /app
## We copy everything in the root directory
## into our /app directory
ADD . /app
## We specify that we now wish to execute 
## any further commands inside our /app
## directory
WORKDIR /app
## we run go build to compile the binary
## executable of our Go program
RUN go build -o fxtract-backend-api .
## Our start command which kicks off
## our newly created binary executable
CMD ["/app/fxtract-backend-api"]


# WORKDIR /src 


# RUN go get github.com/sirupsen/logrus
# RUN go get github.com/streadway/amqp
# RUN go get -u github.com/mitchellh/mapstructure
# RUN go get github.com/gorilla/handlers
# RUN go get -u github.com/prometheus/client_golang
# RUN go get go.mongodb.org/mongo-driver
# RUN go get github.com/gorilla/handlers
# RUN go get -u github.com/WilfredDube/school
# RUN go get github.com/pkg/errors
# RUN go get github.com/teris-io/shortid
# RUN go get github.com/go-redis/redis
# RUN go get gopkg.in/dealancer/validate.v2
# RUN go get github.com/go-redis/redis
# RUN go get go.opentelemetry.io/otel/label github.com/go-redis/redis
# RUN go get go.opentelemetry.io/otel/label
# RUN go get github.com/go-chi/chi
# RUN go get github.com/go-redis/redis
# RUN go get "github.com/lightstep/otel-launcher-go/launcher"
# RUN go get github.com/vmihailenco/msgpack/v5
