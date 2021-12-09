package main

import (
	"context"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/pkedy/golang-dapr/proto/products"
)

const (
	port = ":50151"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedProductsServer
	sync.RWMutex
	products map[string]*pb.Product
}

func newServer() *server {
	return &server{
		products: make(map[string]*pb.Product),
	}
}

func (s *server) GetProduct(ctx context.Context, in *pb.ProductRequest) (*pb.Product, error) {
	log.Println("GetProduct called", in.Id)
	s.RLock()
	defer s.RUnlock()

	product, ok := s.products[in.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "product %q not found", in.Id)
	}

	return product, nil
}

func (s *server) SaveProduct(ctx context.Context, product *pb.Product) (*emptypb.Empty, error) {
	log.Println("SaveProduct called", product)
	s.Lock()
	defer s.Unlock()

	s.products[product.Id] = product

	return &emptypb.Empty{}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterProductsServer(s, newServer())
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
