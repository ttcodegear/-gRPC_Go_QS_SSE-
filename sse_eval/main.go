package main

import (
  "log"
  "net"
  //"io"
  "context"
  //"github.com/golang/protobuf/proto"
  "google.golang.org/grpc"
  "google.golang.org/grpc/metadata"
  //"google.golang.org/grpc/status"
  //"google.golang.org/grpc/codes"
  pb "sse/pb/sse"
)

type ConnectorServer struct {
  pb.UnimplementedConnectorServer
}

func (s *ConnectorServer) EvaluateScript(stream pb.Connector_EvaluateScriptServer) error {
  log.Printf("EvaluateScript")
  header := metadata.Pairs("qlik-cache", "no-store") // Disable caching
  stream.SendHeader(header)

  // Read gRPC metadata
  if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
    entry := md["qlik-scriptrequestheader-bin"][0]
    var header = pb.ScriptRequestHeader{}
    if err := proto.Unmarshal([]byte(entry), &header); err == nil {
      return nil;
    }
  }
  return nil
}

func (s *ConnectorServer) GetCapabilities(ctx context.Context, in *pb.Empty) (*pb.Capabilities, error) {
  log.Printf("GetCapabilities")
  var capabilities = pb.Capabilities {
    AllowScript:      true,
    PluginIdentifier: "Simple SSE Test",
    PluginVersion:    "v0.0.1",
  }
  return &capabilities, nil
}

func main() {
  listen, err := net.Listen("tcp", ":50053")
  if err != nil {
    log.Fatalf("%v", err)
    return
  }
  s := grpc.NewServer()
  var server ConnectorServer
  pb.RegisterConnectorServer(s, &server)
  err = s.Serve(listen)
  if err != nil {
    log.Fatalf("%v", err)
  }
}
