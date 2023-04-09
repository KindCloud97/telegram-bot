package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/KindCloud97/telegram-bot/model"
	"github.com/KindCloud97/telegram-bot/queue"
	proto "github.com/KindCloud97/telegram-bot/usersvc"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"google.golang.org/grpc/metadata"
)

var _ proto.UserServiceServer = (*Service)(nil)

type Service struct {
	proto.UnimplementedUserServiceServer

	bot              *telego.Bot
	queue            *queue.Queue
	chatIDToConnUser map[int64]ConnectedUser
	mu               sync.Mutex
}

type ConnectedUser struct {
	Client proto.UserService_ConnectServer
}

func NewService(bot *telego.Bot, queue *queue.Queue) *Service {
	return &Service{
		bot:              bot,
		queue:            queue,
		chatIDToConnUser: map[int64]ConnectedUser{},
		mu:               sync.Mutex{},
	}
}

// HandleMessage redirects message to operator.
func (s *Service) HandleMessage(m *telego.Message) {
	s.mu.Lock()
	conn, ok := s.chatIDToConnUser[m.Chat.ID]
	s.mu.Unlock()
	if !ok {
		return
	}

	msg := &proto.Message{
		Text: m.Text,
	}

	if len(m.Photo) != 0 {
		// msg.Image = m.Photo[len(m.Photo)-1].FileID	//largest image
		msg.Image = m.Photo[0].FileID //smallest image
	}

	err := conn.Client.Send(msg)
	if err != nil {
		fmt.Println("ERROR! could not send message:", err)
		return
	}
}

func (s *Service) AddToQueue(user model.User) {
	s.queue.Add(user)
}

// GetList get list of all connected users.
func (s *Service) GetList(ctx context.Context, r *proto.GetListRequest) (*proto.GetListResponse, error) {
	users := []*proto.User{}
	list := s.queue.GetAll()

	for id, user := range list {
		users = append(users, &proto.User{
			Id:      id,
			Name:    user.Name,
			Surname: user.Surname,
		})
	}

	return &proto.GetListResponse{
		Users: users,
	}, nil
}

// Connect establishes a bi-directional communication channel.
func (s *Service) Connect(stream proto.UserService_ConnectServer) error {
	userID := metadata.ValueFromIncomingContext(stream.Context(), "id")

	id := strings.Join(userID, "")
	u, ok := s.queue.PopUser(id)
	if !ok {
		return fmt.Errorf("user not found")
	}

	s.mu.Lock()
	s.chatIDToConnUser[u.ChatId] = ConnectedUser{
		Client: stream,
	}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.chatIDToConnUser, u.ChatId)
		s.mu.Unlock()
	}()

	for {
		m, err := stream.Recv()
		if err != nil {
			return err
		}

		_, err = s.bot.SendMessage(tu.Message(tu.ID(u.ChatId), m.Text))
		if err != nil {
			return err
		}
	}
}
