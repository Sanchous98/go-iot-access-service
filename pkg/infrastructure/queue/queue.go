package queue

import "context"

type Queue interface {
	Push(context.Context, string, string)
	Pop(context.Context, string) string
	Subscribe(context.Context, string) <-chan string
	Len(context.Context, string) int
	Clear(context.Context, string)
}
