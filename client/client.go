package main

import (
	chat "assignment_3/chat"
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

var client chat.ChittyChatClient
var wait *sync.WaitGroup
var LamportT int

func init() {
	wait = &sync.WaitGroup{}
	LamportT = 0
}

func connect(user *chat.User) error {
	var streamerror error
	stream, err := client.CreateStream(context.Background(), &chat.Connect{
		User:   user,
		Active: true,
	})
	if err != nil {
		return fmt.Errorf("Connection failed: %v", err)
	}
	wait.Add(1)
	go func(str chat.ChittyChat_CreateStreamClient) {
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

	client = chat.NewChittyChatClient(conn)
	user := &chat.User{
		Name: getUserName(),
	}
	connect(user)
	connectionMSG := &chat.ChatMessage{
		Id:        user.Id,
		Sender:    "Server Message",
		Content:   fmt.Sprintf("Participant %v joined Chitty-Chat", user.Name),
		Timestamp: int32(LamportT),
	}

	client.SendMessage(context.Background(), connectionMSG)
	wait.Add(1)
	go func() {
		defer wait.Done()
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			clockStep()
			msgContent := scanner.Text()
			closeClient := false
			if strings.ToLower(msgContent) == "exit" {
				msgContent = fmt.Sprintf("Participant %v left Chitty-Chat", user.Name)
				closeClient = true
			}
			msg := &chat.ChatMessage{
				Id:        user.Id,
				Sender:    user.Name,
				Content:   msgContent,
				Timestamp: int32(LamportT),
			}
			if len(msg.Content) > 128 {
				fmt.Printf("Message longer than 128 characters\n")

			} else {
				_, err := client.SendMessage(context.Background(), msg)
				if err != nil {
					fmt.Printf("Error sending message: %v\n", err)
					break
				}
			}
			if closeClient {
				os.Exit(0)
			}
		}
	}()

	go func() {
		wait.Wait()
		close(done)
	}()
	<-done
}
