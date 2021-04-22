# gRPC_Go_QS_SSE
$>go verson

go version go1.16.3

https://github.com/protocolbuffers/protobuf/releases/download/v3.15.8/protoc-3.15.8-win64.zip

https://github.com/protocolbuffers/protobuf/releases/download/v3.15.8/protoc-3.15.8-linux-x86_64.zip

$>unzip protoc-xxxx.zip

$>go env -w GO111MODULE=on

$>go get -u github.com/golang/protobuf/protoc-gen-go





$>mkdir simple

$>cd simple

$>emacs greet.proto

[greet.proto]

-----

...

option go_package = "pb/greeter_server";

...

-----

[Protocol Buffers - Go Reference - Go Generated Code - Packages]

https://developers.google.com/protocol-buffers/docs/reference/go-generated#package

$>protoc --proto_path=./ --go_out=plugins=grpc:./ greet.proto

$>ls ./pb/greeter_server/greet.pb.go

$>go mod init simple

$>go get -u google.golang.org/grpc

$>go get -u google.golang.org/protobuf/reflect/protoreflect

$>go get -u google.golang.org/protobuf/runtime/protoimpl

$>mkdir greeter_server

$>emacs greeter_server/main.go

$>go run greeter_server/main.go

$>mkdir greeter_client

$>emacs greeter_client/main.go

$>go run greeter_client/main.go





$>mkdir simple_ssl

$>cd simple_ssl

$>openssl req -newkey rsa:2048 -nodes -keyout server.key -x509 -days 3650 -out server.crt

$>...

$>go mod init simple_ssl

$>...

[greeter_server/main.go]

-----

import (

  ...

  "google.golang.org/grpc/credentials"

  pb "simple_ssl/pb/greeter_server"

)

...

  serverCert, err := credentials.NewServerTLSFromFile("server.crt", "server.key")

  s := grpc.NewServer(grpc.Creds(serverCert))

  ...

-----

[greeter_client/main.go]

-----

import (

  ...

  "google.golang.org/grpc/credentials"

  "crypto/tls"

  pb "simple_ssl/pb/greeter_server"

)

...

  creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})

  conn, err := grpc.Dial("jatok-tts1:50051", grpc.WithTransportCredentials(creds))

  ...

-----

$>...





$>mkdir stream

$>cd stream

$>emacs hellostreamingworld.proto

[hellostreamingworld.proto]

-----

...

option go_package = "pb/greeter_server";

...

-----

$>protoc --proto_path=./ --go_out=plugins=grpc:./ hellostreamingworld.proto

$>ls ./pb/greeter_server/hellostreamingworld.pb.go

$>go mod init stream

$>...

[greeter_server/main.go]

-----

import (

  ...

  pb "stream/pb/greeter_server"

)

...

func (s *MultiGreeterServer) SayHello(stream pb.MultiGreeter_SayHelloServer) error {

  ...

}

...

-----

[greeter_client/main.go]

-----

import (

  ...

  pb "stream/pb/greeter_server"

)

...

  stream, err := client.SayHello(context.Background())

  ...

  stream.CloseSend()

  for {

    reply, err := stream.Recv()

    ...

-----

$>...





$>mkdir sse

$>cd sse

[Protocol Buffers - Go Reference - Go Generated Code - Packages]

https://developers.google.com/protocol-buffers/docs/reference/go-generated#package

$>protoc --proto_path=./ --go_opt=MServerSideExtension.proto=pb/sse --go_out=plugins=grpc:./ ServerSideExtension.proto

$>ls ./pb/sse/ServerSideExtension.pb.go

$>go mod init sse

$>go get -u google.golang.org/grpc

$>go get -u google.golang.org/protobuf/reflect/protoreflect

$>go get -u google.golang.org/protobuf/runtime/protoimpl

$>mkdir SSE_Example

$>emacs SSE_Example/main.go

[SSE_Example/main.go]

-----

import (

  ...

  "io"

  "context"

  "github.com/golang/protobuf/proto"

  "google.golang.org/grpc"

  "google.golang.org/grpc/metadata"

  "google.golang.org/grpc/status"

  "google.golang.org/grpc/codes"

  pb "sse/pb/sse"

)

...

func (s *ConnectorServer) ExecuteFunction(stream pb.Connector_ExecuteFunctionServer) error {

  ...

}



func (s *ConnectorServer) GetCapabilities(ctx context.Context, in *pb.Empty) (*pb.Capabilities, error) {

  ...

}

...

-----

$>go run SSE_Example/main.go



C:\Users\[user]\Documents\Qlik\Sense\Settings.ini

------

[Settings 7]

SSEPlugin=Column,localhost:50053



------





$>mkdir sse_eval

$>cd sse_eval

...

