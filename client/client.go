package main

import (
	chat "assignment_3/chat"
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

var client chat.BroadcastClient
var wait *sync.WaitGroup

func init() {
	wait = &sync.WaitGroup{}
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
	go func(str chat.Broadcast_CreateStreamClient) {
		defer wait.Done()

		for {
			msg, err := str.Recv()
			if err != nil {
				streamerror = fmt.Errorf("error reading message: %v", err)
				break
			}
			fmt.Printf("Sender: %v\n %v\n", msg.Sender, msg.Content)
		}
	}(stream)
	return streamerror
}

func main() {
	timestamp := time.Now().UTC()
	done := make(chan int)
	name := flag.String("N", "anon", "name of the user")
	flag.Parse()
	id := sha256.Sum256([]byte(timestamp.String() + *name))

	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}

	client = chat.NewBroadcastClient(conn)
	user := &chat.User{
		Id:   hex.EncodeToString(id[:]),
		Name: *name,
	}

	connect(user)
	wait.Add(1)
	go func() {
		defer wait.Done()

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			msg := &chat.Message{
				Id:        user.Id,
				Sender:    user.Name,
				Content:   scanner.Text(),
				Timestamp: timestamp.String(),
			}

			_, err := client.BroadcastMessage(context.Background(), msg)
			if err != nil {
				fmt.Printf("Error sending message: %v\n", err)
				break
			}
		}
	}()

	go func() {
		wait.Wait()
		close(done)
	}()
	<-done
}
