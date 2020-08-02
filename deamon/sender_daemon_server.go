package deamon

import (
	context "context"
	"fmt"
	"io"
	"net"
	"sharef/api"
	"sharef/streamer"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// UnimplementedSenderServer can be embedded to have forward compatible implementations.
type SenderDaemonServer struct {
	sender *streamer.Sender
}

func StartSenderDaemonServer(sender *streamer.Sender, port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	api.RegisterSenderServer(grpcServer, &SenderDaemonServer{sender: sender})
	// determine whether to use TLS
	return grpcServer.Serve(lis)
}

func (*SenderDaemonServer) Hello(context.Context, *api.HelloRequest) (*api.HelloReply, error) {
	return &api.HelloReply{}, nil
}
func (s *SenderDaemonServer) SendFile(req *api.SendFileRequest, stream api.Sender_SendFileServer) error {
	streamer, err := s.sender.InitFileStreamer(req.Filename)
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}

	// reader := bytes.NewBuffer([]byte{})
	// writer := bytes.NewBuffer([]byte{})
	reader, writer := io.Pipe()
	streamer.SetOutput(writer)

	streamer.Stream()
	// defer streamer.Close()
	for {
		select {
		case <-streamer.Done:
			return nil
		default:
		}

		data := make([]byte, 4096)
		n, err := reader.Read(data)
		if err != nil && err != io.EOF {
			return status.Errorf(codes.Internal, err.Error())
		}

		out := &api.STDOutput{
			Line: string(data[:n]),
		}

		if err := stream.Send(out); err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}
	}
}
