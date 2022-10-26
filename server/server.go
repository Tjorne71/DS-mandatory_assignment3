package main

import (
	"context"
	"log"
	"net"
	"os"
	"sync"

	chat "assignment_3/chat"

	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
)

var grpclog glog.LoggerV2

var userMap = make(map[string]int)

func init() {
	grpclog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Connection struct {
	stream   chat.ChittyChat_CreateStreamServer
	userName string
	active   bool
	error    chan error
}

type ChittyChatServer struct {
	Connection []*Connection
}

func (s *ChittyChatServer) CreateStream(pconn *chat.Connect, stream chat.ChittyChat_CreateStreamServer) error {
	connection := &Connection{
		stream:   stream,
		userName: pconn.User.Name,
		active:   true,
		error:    make(chan error),
	}

	s.Connection = append(s.Connection, connection)
	return <-connection.error
}

func (s *ChittyChatServer) SendMessage(ctx context.Context, msg *chat.ChatMessage) (*chat.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)
	for _, conn := range s.Connection {
		wait.Add(1)

		go func(msg *chat.ChatMessage, conn *Connection) {
			defer wait.Done()
			if conn.active {
				err := conn.stream.Send(msg)
				grpclog.Info("Sending msg to: ", conn.stream)
				if err != nil {
					conn.active = false
					conn.error <- err
				}
			}
		}(msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()
	<-done
	return &chat.Close{}, nil
}

func main() {
	var connections []*Connection

	server := &ChittyChatServer{connections}

	grpcServer := grpc.NewServer()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("error creating the server: %v", err)
	}

	grpclog.Info("Starting server at port 8080")

	chat.RegisterChittyChatServer(grpcServer, server)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("error serving: %v", err)
	}
}
