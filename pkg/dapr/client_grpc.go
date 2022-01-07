package dapr

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	v1 "github.com/dapr/dapr/pkg/proto/common/v1"
	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/pkedy/golang-dapr/pkg/components/secrets"
	"github.com/pkedy/golang-dapr/pkg/components/state"
	"github.com/pkedy/golang-dapr/pkg/errorz"
)

type GRPC struct {
	client pb.DaprClient
}

var (
	GRPCADDRESS = fmt.Sprintf("127.0.0.1:%s", os.Getenv("DAPR_GRPC_PORT"))

	_ = state.Store((*HTTP)(nil))
	_ = secrets.Store((*HTTP)(nil))
)

func NewGRPC(ctx context.Context) (*GRPC, error) {
	conn, err := grpc.DialContext(
		ctx,
		GRPCADDRESS,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(UnaryClientInterceptor),
	)
	if err != nil {
		return nil, err
	}
	client := pb.NewDaprClient(conn)
	return &GRPC{
		client: client,
	}, nil
}

func (c *GRPC) Name() string {
	return "Custom gRPC"
}

func (c *GRPC) SetState(ctx context.Context, store string, items ...state.Item) error {
	stateItems := make([]*v1.StateItem, len(items))
	for i := range items {
		item := items[i]
		data, err := json.Marshal(item.Value)
		if err != nil {
			return errorz.Internal(err, "could not serialize value for key %q", item.Key)
		}
		stateItems[i] = &v1.StateItem{
			Key:   item.Key,
			Etag:  etagGRPC(item.ETag),
			Value: data,
		}
	}
	c.client.SaveState(ctx, &pb.SaveStateRequest{
		StoreName: store,
		States:    stateItems,
	})
	return nil
}

func (c *GRPC) GetState(ctx context.Context, store string, key string, target interface{}) error {
	state, err := c.client.GetState(ctx, &pb.GetStateRequest{
		StoreName:   store,
		Key:         key,
		Consistency: v1.StateOptions_CONSISTENCY_STRONG,
	})
	if err != nil {
		return errorz.Internal(err, "could not load state %q", key)
	}
	if state.Data == nil {
		return errorz.NotFound("key %q not found", key)
	}
	if err = json.Unmarshal(state.Data, target); err != nil {
		return errorz.Internal(err, "could decode state %q", key)
	}
	return nil
}

func (c *GRPC) GetSecret(ctx context.Context, store string, name string, target interface{}) error {
	secret, err := c.client.GetSecret(ctx, &pb.GetSecretRequest{
		StoreName: store,
		Key:       name,
	})
	if err != nil {
		return errorz.Internal(err, "could not load secret %q", name)
	}
	if secret.Data == nil {
		return errorz.NotFound("secret %q not found", name)
	}
	dataBytes, err := json.Marshal(secret.Data)
	if err != nil {
		return errorz.Internal(err, "could decode secret %q", name)
	}
	err = json.Unmarshal(dataBytes, target)
	if err != nil {
		return errorz.Internal(err, "could decode secret %q", name)
	}
	return nil
}

func etagGRPC(value string) *v1.Etag {
	if value == "" {
		return nil
	}
	return &v1.Etag{
		Value: value,
	}
}

// UnaryClientInterceptor for passing incoming metadata to outgoing metadata
func UnaryClientInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {
	// Take the incoming metadata and transfer it to the outgoing metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// InvokingContext returns a new context with the target Dapr App ID added to outgoing metadata.
func InvokingContext(ctx context.Context, daprAppID string) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	} else {
		md = md.Copy() // Make a copy for concurrency reasons.
	}
	md.Append("dapr-app-id", daprAppID)

	return metadata.NewOutgoingContext(ctx, md)
}
