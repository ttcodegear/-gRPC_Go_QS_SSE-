# gRPC_Go_QS_SSE

----- For latest Go and protoc/gRPC verion -----

$>go version

go version go1.17

https://github.com/protocolbuffers/protobuf/releases/download/v3.17.3/protoc-3.17.3-win64.zip

https://github.com/protocolbuffers/protobuf/releases/download/v3.17.3/protoc-3.17.3-linux-x86_64.zip

$>unzip protoc-xxxx.zip

$>go env -w GO111MODULE=on

$>go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

$>go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest



$>mkdir simple2

$>cd simple2

$>emacs greet.proto

[greet.proto]

-----

...

option go_package = "pb/greeter_server";

...

-----

[gRPC - Documentation - Languages - Go - Quick start - Quick start - Regenerate gRPC code]

https://grpc.io/docs/languages/go/quickstart/#regenerate-grpc-code

$>mkdir -p pb/greeter_server

$>protoc --go_out=./pb/greeter_server --go_opt=paths=source_relative --go-grpc_out=./pb/greeter_server --go-grpc_opt=paths=source_relative greet.proto

$>ls ./pb/greeter_server/greet.pb.go

$>ls ./pb/greeter_server/greet_grpc.pb.go

$>go mod init simple2

$>go get -u google.golang.org/grpc

$>mkdir greeter_server

$>emacs greeter_server/main.go

[greeter_server/main.go]

-----

...

pb "simple2/pb/greeter_server"

...

-----

$>go run greeter_server/main.go

$>mkdir greeter_client

$>emacs greeter_client/main.go

[greeter_client/main.go]

-----

...

pb "simple2/pb/greeter_server"

...

-----

$>go run greeter_client/main.go

--------------------------------------------


----- For older Go and protoc/gRPC verion -----

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

--------------------------------------------





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

$>go get -u github.com/ahmetb/go-linq/v3

$>go get -u github.com/deckarep/golang-set

...

[for SSL]

------

...

  "google.golang.org/grpc/credentials"

...

  serverCert, err := credentials.NewServerTLSFromFile("SSE_Example/sse_server_cert.pem", "SSE_Example/sse_server_key.pem")

  if err != nil {

    log.Fatalf("%v", err)

    return

  }

  s := grpc.NewServer(grpc.Creds(serverCert))

...

------

C:\Users\[user]\Documents\Qlik\Sense\Settings.ini

------

[Settings 7]

SSEPlugin=Column,localhost:50053,C:\...\sse_Column_generated_certs\sse_Column_client_certs_used_by_qlik



------

