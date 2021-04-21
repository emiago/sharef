package deamon

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/emiraganov/sharef/api"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type SenderDeamonClient struct {
	conn   *grpc.ClientConn
	client api.SenderClient
}

func InitSenderDeamonClient() *SenderDeamonClient {
	conn, err := grpc.Dial(":9876", grpc.WithInsecure())
	if err != nil {
		return nil
	}

	client := api.NewSenderClient(conn)

	_, err = client.Hello(context.Background(), &api.HelloRequest{})
	if err != nil {
		return nil
	}

	s := &SenderDeamonClient{
		conn:   conn,
		client: client,
	}
	return s
}

func (s *SenderDeamonClient) ProcessArgs(args []string) {
	for _, file := range args {
		file, err := filepath.Abs(file)
		if err != nil {
			log.Fatal(err)
		}

		out, err := s.client.SendFile(context.Background(), &api.SendFileRequest{Filename: file})
		if err != nil {
			log.Fatal(err)
		}

		for {
			stdout, err := out.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatal(err)
			}
			fmt.Print(stdout.Line)
		}
	}
}

func (s *SenderDeamonClient) Close() error {
	return s.conn.Close()
}
