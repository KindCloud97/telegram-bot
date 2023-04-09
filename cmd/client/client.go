package main

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"time"

	proto "github.com/KindCloud97/telegram-bot/usersvc"
	"github.com/dolmen-go/kittyimg"
	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	botToken := os.Getenv("TOKEN")
	bot, err := telego.NewBot(botToken)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ctx := context.Background()
	conn, err := grpc.Dial("127.0.0.1:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	checkError(err)

	client := proto.NewUserServiceClient(conn)

	var resp *proto.GetListResponse
	for {
		resp, err = client.GetList(ctx, &proto.GetListRequest{})
		checkError(err)
		if len(resp.Users) != 0 {
			break
		}

		time.Sleep(time.Second)
	}

	user := resp.Users[0]
	ctx = metadata.AppendToOutgoingContext(ctx, "id", user.Id)
	stream, err := client.Connect(ctx)
	checkError(err)

	fmt.Printf("Connected to chat with %s!\n", user.Name)
	err = stream.Send(&proto.Message{
		Text:  fmt.Sprintf("Hello %s!\nHow can I help you?\n", user.Name),
		Image: "",
	})
	checkError(err)

	go func() {
		for {
			m, err := stream.Recv()
			checkError(err)

			fmt.Printf("[%s %s]: %s\n", user.Name, user.Surname, m.Text)

			if m.Image != "" {
				f, err := bot.GetFile(&telego.GetFileParams{
					FileID: m.Image,
				})
				checkError(err)

				img, err := getimage(f.FilePath, botToken)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return
				}

				displayImage(img)
			}
		}
	}()

	fmt.Print("[You]: ")
	for {
		var line string

		fmt.Scanln(&line)
		err := stream.Send(&proto.Message{
			Text:  line,
			Image: "",
		})
		checkError(err)
	}
}

func getimage(filepath string, botToken string) (image.Image, error) {
	resp, err := http.Get("https://api.telegram.org/file/bot" + botToken + "/" + filepath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return img, err
}

// displayImage renders an image to the terminal.
func displayImage(m image.Image) {
	kittyimg.Fprintln(os.Stdout, m)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
