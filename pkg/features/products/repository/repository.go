package repository

import (
	"context"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/pkedy/golang-dapr/pkg/errorz"
	"github.com/pkedy/golang-dapr/pkg/features/products"
	pb "github.com/pkedy/golang-dapr/proto/products"
)

var (
	GRPCADDRESS = fmt.Sprintf("127.0.0.1:%s", os.Getenv("DAPR_GRPC_PORT"))
)

type Repository struct {
	log    logr.Logger
	conn   *grpc.ClientConn
	client pb.ProductsClient
}

func New(log logr.Logger) (*Repository, error) {
	conn, err := grpc.Dial(GRPCADDRESS, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("could not connect: %v", err)
	}
	client := pb.NewProductsClient(conn)

	return &Repository{
		log:    log,
		conn:   conn,
		client: client,
	}, nil
}

func (r *Repository) Close() error {
	return r.conn.Close()
}

func (r *Repository) Save(ctx context.Context, product *products.Product) error {
	r.log.Info("Invoking products service: SaveProduct")
	ctx = metadata.AppendToOutgoingContext(ctx, "dapr-app-id", "products")
	_, err := r.client.SaveProduct(ctx, &pb.Product{
		Id:          product.ID,
		Description: product.Description,
		Price:       product.Price,
	})
	return err
}

func (r *Repository) Load(ctx context.Context, id string) (*products.Product, error) {
	r.log.Info("Invoking products service: GetProduct")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	md.Append("dapr-app-id", "products")
	ctx = metadata.NewOutgoingContext(ctx, md)
	//AppendToOutgoingContext(ctx, "dapr-app-id", "products")
	product, err := r.client.GetProduct(ctx, &pb.ProductRequest{Id: id})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, errorz.NotFound(st.Message())
		}
		return nil, err
	}
	return &products.Product{
		ID:          product.Id,
		Description: product.Description,
		Price:       product.Price,
	}, nil
}
