package services

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/queue"
	"bitbucket.org/4suites/iot-service-golang/pkg/utils"
	"context"
	"github.com/goccy/go-json"
)

const commandsQueue = "commands"

type Command struct {
	CommandId   int
	Queue       string
	CommandName string
	Payload     map[string]any
}

type CommandHandler interface {
	CanHandle(Command) bool
	Handle(Command)
}

type CommandQueue struct {
	queue    queue.Queue      `inject:""`
	handlers []CommandHandler `inject:"iot.command_handler"`
}

func (q *CommandQueue) Launch(ctx context.Context) {
	var command Command

	select {
	case item := <-q.queue.Subscribe(ctx, commandsQueue):
		_ = json.Unmarshal(utils.StrToBytes(item), &command)

		for _, handler := range q.handlers {
			if handler.CanHandle(command) {
				handler.Handle(command)
				break
			}
		}
	}
}
