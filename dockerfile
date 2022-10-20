FROM golang:alpine as build-env

ENV GO111MODULE=on

RUN apk update && apk add bash ca-certificates git gcc g++ libc-dev

RUN mkdir /ds-mandatory_assignment3
RUN mkdir -p /ds-mandatory_assignment3/chat  

WORKDIR /ds-mandatory_assignment3

COPY ./chat/chat.pb.go /ds-mandatory_assignment3/chat
COPY ./chat/chat_grpc.pb.go /ds-mandatory_assignment3/chat
COPY ./server/server.go /ds-mandatory_assignment3

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go build -o ds-mandatory_assignment3 .

CMD ./ds-mandatory_assignment3

# docker build --tag=ds-mandatory_assignment3 .  
# docker run -it -p 8080:8080 ds-mandatory_assignment3