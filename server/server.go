package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	chat "assignment_3/chat"

	"google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/stats"
)

var grpclog glog.LoggerV2

var userMap = make(map[string]int)

func init() {
	grpclog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type Handler struct {
}

func (h *Handler) TagRPC(context.Context, *stats.RPCTagInfo) context.Context {
	log.Println("TagRPC")
	return context.Background()
}

// HandleRPC processes the RPC stats.
func (h *Handler) HandleRPC(context.Context, stats.RPCStats) {
	log.Println("HandleRPC")
}

func (h *Handler) TagConn(context.Context, *stats.ConnTagInfo) context.Context {

	log.Println("Tag Conn")
	return context.Background()
}

// HandleConn processes the Conn stats.
func (h *Handler) HandleConn(c context.Context, s stats.ConnStats) {
	switch s.(type) {
	case *stats.ConnEnd:
		log.Println("get connEnd")
		fmt.Printf("client %v disconnected", c.Value("name"))
		break
	}
}

type Connection struct {
	stream chat.Broadcast_CreateStreamServer
	userName     string
	active bool
	error  chan error
}

type Server struct {
	Connection []*Connection
}

func (s *Server) CreateStream(pconn *chat.Connect, stream chat.Broadcast_CreateStreamServer) error {
	connection := &Connection{
		stream: stream,
		userName:     pconn.User.Name,
		active: true,
		error:  make(chan error),
	}

	s.Connection = append(s.Connection, connection)
	return <-connection.error
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *chat.Message) (*chat.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)
	for _, conn := range s.Connection {
		wait.Add(1)

		go func(msg *chat.Message, conn *Connection) {
			defer wait.Done()
			if conn.active {
				err := conn.stream.Send(msg)
				grpclog.Info("Sending msg to: ", conn.stream)
				if err != nil {
					grpclog.Errorf("error sending msg on stream: %v - Error: %v", conn.stream, err)
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

	server := &Server{connections}

	grpcServer := grpc.NewServer(grpc.StatsHandler(&Handler{}))

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("error creating the server: %v", err)
	}

	grpclog.Info("Starting server at port 8080")

	chat.RegisterBroadcastServer(grpcServer, server)
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("error serving: %v", err)
	}
}


