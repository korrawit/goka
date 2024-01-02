package collector

import (
	"context"
	"encoding/json"
	"io"

	"github.com/lovoo/goka"
	"github.com/lovoo/goka/codec"
	messaging "github.com/lovoo/goka/examples/3-messaging"
	"github.com/lovoo/goka/examples/3-messaging/topicinit"
)

const maxMessages = 5

var (
	group goka.Group = "collector"
	Table goka.Table = goka.GroupTable(group)
)

type MessageListCodec struct{}

func (c *MessageListCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (c *MessageListCodec) Decode(data []byte) (interface{}, error) {
	var m []messaging.Message
	err := json.Unmarshal(data, &m)
	return m, err
}

func (c *MessageListCodec) DecodeP(data []byte) (interface{}, io.Closer, error) {
	dec, err := c.Decode(data)
	return dec, codec.NoopCloser, err
}

func collect(ctx goka.Context, msg interface{}) {
	var ml []messaging.Message
	if v := ctx.Value(); v != nil {
		ml = v.([]messaging.Message)
	}

	m := msg.(*messaging.Message)
	ml = append(ml, *m)

	if len(ml) > maxMessages {
		ml = ml[len(ml)-maxMessages:]
	}
	ctx.SetValue(ml)
}

func PrepareTopics(brokers []string) {
	topicinit.EnsureStreamExists(string(messaging.ReceivedStream), brokers)
}

func Run(ctx context.Context, brokers []string) func() error {
	return func() error {
		g := goka.DefineGroup(group,
			goka.Input(messaging.ReceivedStream, new(messaging.MessageCodec), collect),
			goka.Persist(new(MessageListCodec)),
		)
		p, err := goka.NewProcessor(brokers, g)
		if err != nil {
			return err
		}
		return p.Run(ctx)
	}
}
