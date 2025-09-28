FROM golang:1.24.5-alpine 

WORKDIR /app


 COPY go.mod go.sum ./

 RUN go mod download && go mod verify

COPY . .


RUN go build -o cmd/main cmd/main.go

CMD ["/app/cmd/main"]

