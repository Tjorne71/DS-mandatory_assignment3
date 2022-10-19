FROM golang:alpine as build-env
ENV GO111MODULE=on

RUN apk update && apk add bash ca-certificates git gcc g++ libc-dev

RUN mkdir /DS-MANDATORY_ASSIGNMENTS3
RUN mkdir -p /DS-MANDATORY_ASSIGNMENTS3/chat 
RUN mkdir -p /DS-MANDATORY_ASSIGNMENTS3/client 
RUN mkdir -p /DS-MANDATORY_ASSIGNMENTS3/server 

WORKDIR /DS-MANDATORY_ASSIGNMENTS3

COPY ./chat/chat.pb.go /DS-MANDATORY_ASSIGNMENTS3/chat
COPY ./chat/chat_grpc.pb.go /DS-MANDATORY_ASSIGNMENTS3/chat

COPY ./server/server.go /DS-MANDATORY_ASSIGNMENTS3/server 
COPY ./client/client.go /DS-MANDATORY_ASSIGNMENTS3/client 

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go build -o DS-MANDATORY_ASSIGNMENTS3 .

CMD ./ds-mandatory_assignment3