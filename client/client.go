package main

import (
	chat "assignment_3/chat"
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"sync"

	"google.golang.org/grpc"
)

var client chat.BroadcastClient
var wait *sync.WaitGroup
var LamportT int

func init() {
	wait = &sync.WaitGroup{}
	LamportT = 0
}

func ctxSetName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, "name", name)
}

func connect(user *chat.User) error {
	var streamerror error
	ctx := context.Background()
	ctx = ctxSetName(ctx, user.Name)
	stream, err := client.CreateStream(ctx, &chat.Connect{
		User:   user,
		Active: true,
	})
	if err != nil {
		return fmt.Errorf("Connection failed: %v", err)
	}
	wait.Add(1)
	go func(str chat.Broadcast_CreateStreamClient) {
		defer wait.Done()
		if err != nil {
			streamerror = fmt.Errorf("error reading message: %v", err)
		}
		for {
			msg, err := str.Recv()
			if err != nil {
				streamerror = fmt.Errorf("error reading message: %v", err)
				break
			}
			pickStep(int(msg.Timestamp))
			fmt.Printf("Sender: %v\n\t%v\nLamport timestamp: %d\n", msg.Sender, msg.Content, LamportT)
		}
	}(stream)
	return streamerror
}

func clockStep() {
	LamportT++
}

func pickStep(tprime int) {
	LamportT = int(math.Max(float64(LamportT), float64(tprime))) + 1
}

func getUserName() string {
	fmt.Printf("Input Your User Name: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func main() {
	done := make(chan int)
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Error Dialling Host: %v\n", err)
	}

	client = chat.NewBroadcastClient(conn)
	user := &chat.User{
		Name: getUserName(),
	}
	connect(user)
	connectionMSG := &chat.Message{
		Id:        user.Id,
		Sender:    "Server Message",
		Content:   fmt.Sprintf("Participant %v joined Chitty-Chat", user.Name),
		Timestamp: int32(LamportT),
	}

	client.BroadcastMessage(context.Background(), connectionMSG)
	wait.Add(1)
	go func() {
		defer wait.Done()
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			clockStep()
			msg := &chat.Message{
				Id:        user.Id,
				Sender:    user.Name,
				Content:   scanner.Text(),
				Timestamp: int32(LamportT),
			}
			if len(msg.Content) > 128 {
				fmt.Printf("Message longer than 128 characters\n")

			} else {
				_, err := client.BroadcastMessage(context.Background(), msg)
				if err != nil {
					fmt.Printf("Error sending message: %v\n", err)
					break
				}
			}

		}
	}()

	go func() {
		wait.Wait()
		close(done)
	}()
	<-done
}
