package list

import (
	"context"
	"github.com/atomix/atomix-go-client/pkg/client/primitive"
	"github.com/atomix/atomix-go-client/pkg/client/session"
	"github.com/atomix/atomix-go-client/pkg/client/util"
	pb "github.com/atomix/atomix-go-client/proto/atomix/list"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	"io"
)

type ListClient interface {
	GetList(ctx context.Context, name string, opts ...session.SessionOption) (List, error)
}

type List interface {
	primitive.Primitive
	Append(ctx context.Context, value string) error
	Insert(ctx context.Context, index int, value string) error
	Get(ctx context.Context, index int) (string, error)
	Remove(ctx context.Context, index int) (string, error)
	Size(ctx context.Context) (int, error)
	Items(ctx context.Context, ch chan<- string) error
	Listen(ctx context.Context, ch chan<- *ListEvent) error
	Clear(ctx context.Context) error
}

type ListEventType string

const (
	EventInserted ListEventType = "added"
	EventRemoved  ListEventType = "removed"
)

type ListEvent struct {
	Type  ListEventType
	Value string
}

func New(ctx context.Context, name primitive.Name, partitions []*grpc.ClientConn, opts ...session.SessionOption) (List, error) {
	i, err := util.GetPartitionIndex(name.Name, len(partitions))
	if err != nil {
		return nil, err
	}
	return newList(ctx, name, partitions[i], opts...)
}

func newList(ctx context.Context, name primitive.Name, conn *grpc.ClientConn, opts ...session.SessionOption) (*list, error) {
	client := pb.NewListServiceClient(conn)
	sess, err := session.New(ctx, name, &SessionHandler{client: client}, opts...)
	if err != nil {
		return nil, err
	}
	return &list{
		name:    name,
		client:  client,
		session: sess,
	}, nil
}

type list struct {
	name    primitive.Name
	client  pb.ListServiceClient
	session *session.Session
}

func (l *list) Name() primitive.Name {
	return l.name
}

func (l *list) Append(ctx context.Context, value string) error {
	request := &pb.AppendRequest{
		Header: l.session.NextHeader(),
		Value:  value,
	}

	_, err := l.client.Append(ctx, request)
	return err
}

func (l *list) Insert(ctx context.Context, index int, value string) error {
	request := &pb.InsertRequest{
		Header: l.session.NextHeader(),
		Index:  uint32(index),
		Value:  value,
	}

	_, err := l.client.Insert(ctx, request)
	return err
}

func (l *list) Get(ctx context.Context, index int) (string, error) {
	request := &pb.GetRequest{
		Header: l.session.GetHeader(),
		Index:  uint32(index),
	}

	response, err := l.client.Get(ctx, request)
	if err != nil {
		return "", err
	}
	return response.Value, nil
}

func (l *list) Remove(ctx context.Context, index int) (string, error) {
	request := &pb.RemoveRequest{
		Header: l.session.NextHeader(),
		Index:  uint32(index),
	}

	response, err := l.client.Remove(ctx, request)
	if err != nil {
		return "", err
	}
	return response.Value, nil
}

func (l *list) Size(ctx context.Context) (int, error) {
	request := &pb.SizeRequest{
		Header: l.session.GetHeader(),
	}

	response, err := l.client.Size(ctx, request)
	if err != nil {
		return 0, err
	}
	return int(response.Size), nil
}

func (l *list) Items(ctx context.Context, ch chan<- string) error {
	request := &pb.IterateRequest{
		Header: l.session.GetHeader(),
	}
	entries, err := l.client.Iterate(ctx, request)
	if err != nil {
		return err
	}

	go func() {
		for {
			response, err := entries.Recv()
			if err == io.EOF {
				close(ch)
				break
			}

			if err != nil {
				glog.Error("Failed to receive items stream", err)
			}
			ch <- response.Value
		}
	}()
	return nil
}

func (l *list) Listen(ctx context.Context, c chan<- *ListEvent) error {
	request := &pb.EventRequest{
		Header: l.session.NextHeader(),
	}
	events, err := l.client.Listen(ctx, request)
	if err != nil {
		return err
	}

	go func() {
		for {
			response, err := events.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				glog.Error("Failed to receive event stream", err)
			}

			var t ListEventType
			switch response.Type {
			case pb.EventResponse_ADDED:
				t = EventInserted
			case pb.EventResponse_REMOVED:
				t = EventRemoved
			}

			// If no stream headers are provided by the server, immediately complete the event.
			if len(response.Header.Streams) == 0 {
				c <- &ListEvent{
					Type:  t,
					Value: response.Value,
				}
			} else {
				// Wait for the stream to advanced at least to the responses.
				stream := response.Header.Streams[0]
				_, ok := <-l.session.WaitStream(stream)
				if ok {
					c <- &ListEvent{
						Type:  t,
						Value: response.Value,
					}
				}
			}
		}
	}()
	return nil
}

func (l *list) Clear(ctx context.Context) error {
	request := &pb.ClearRequest{
		Header: l.session.NextHeader(),
	}
	_, err := l.client.Clear(ctx, request)
	return err
}

func (l *list) Close() error {
	return l.session.Close()
}

func (l *list) Delete() error {
	return l.session.Delete()
}