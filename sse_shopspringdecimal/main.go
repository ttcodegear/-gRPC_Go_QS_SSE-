package main

import (
  "log"
  "net"
  "io"
  "context"
  "github.com/golang/protobuf/proto"
  "google.golang.org/grpc"
  "google.golang.org/grpc/metadata"
  "google.golang.org/grpc/status"
  "google.golang.org/grpc/codes"
  pb "sse/pb/sse"
  "github.com/shopspring/decimal"
)

type ConnectorServer struct {
  pb.UnimplementedConnectorServer
}

func (s *ConnectorServer) BigSum(stream pb.Connector_ExecuteFunctionServer) error {
  log.Printf("BigSum")
  var params []string
  for {
    in, err := stream.Recv()
    if err == io.EOF {
      result, _ := decimal.NewFromString("0")
      for _, str := range params {
        num, _ := decimal.NewFromString(str)
        result = result.Add(num) // Col1 + Col1 + ...
      }
      log.Printf(result.String())
      dual := pb.Dual{ StrData: result.String() }
      row := pb.Row{ Duals: []*pb.Dual{ &dual } }
      var reply pb.BundledRows
      reply.Rows = append(reply.Rows, &row)
      if err := stream.Send(&reply); err != nil {
         return err
      }
      return nil
    }
    if err != nil {
      return err
    }
    for _, row := range in.Rows {
      params = append(params, row.Duals[0].StrData) // row=[Col1]
    }
  }
}

func (s *ConnectorServer) BigAdd(stream pb.Connector_ExecuteFunctionServer) error {
  log.Printf("BigAdd")
  for {
    in, err := stream.Recv()
    if err == io.EOF {
      return nil
    }
    if err != nil {
      return err
    }
    var response_rows pb.BundledRows
    for _, row := range in.Rows {
      param1, _ := decimal.NewFromString(row.Duals[0].StrData) // row=[Col1,Col2]
      param2, _ := decimal.NewFromString(row.Duals[1].StrData)
      result := param1.Add(param2)      // Col1 + Col2
      log.Printf(result.String())
      dual := pb.Dual{ StrData: result.String() }
      r := pb.Row{ Duals: []*pb.Dual{ &dual } }
      response_rows.Rows = append(response_rows.Rows, &r)
    }
    if err := stream.Send(&response_rows); err != nil {
       return err
    }
  }
}

func (s *ConnectorServer) GetFunctionId(stream pb.Connector_ExecuteFunctionServer) int32 {
  // Read gRPC metadata
  if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
    entry := md["qlik-functionrequestheader-bin"][0]
    var header = pb.FunctionRequestHeader{}
    if err := proto.Unmarshal([]byte(entry), &header); err == nil {
      return header.FunctionId
    }
  }
  return int32(-1)
}

func (s *ConnectorServer) ExecuteFunction(stream pb.Connector_ExecuteFunctionServer) error {
  log.Printf("ExecuteFunction")
  header := metadata.Pairs("qlik-cache", "no-store") // Disable caching
  stream.SendHeader(header)
  
  func_id := s.GetFunctionId(stream)
  if func_id == 0 {
    return s.BigSum(stream)
  } else if(func_id == 1) {
    return s.BigAdd(stream)
  } else {
    return status.Errorf(codes.Unimplemented, "Method not implemented!")
  }
}

func (s *ConnectorServer) GetCapabilities(ctx context.Context, in *pb.Empty) (*pb.Capabilities, error) {
  log.Printf("GetCapabilities")
  var capabilities = pb.Capabilities {
    AllowScript:      false,
    PluginIdentifier: "Simple SSE Test",
    PluginVersion:    "v0.0.1",
  }
  
  // BigSum
  var func0 = pb.FunctionDefinition {
    FunctionId:   0,                           // 関数ID
    Name:         "BigSum",                    // 関数名
    FunctionType: pb.FunctionType_AGGREGATION, // 関数タイプ=0=スカラー,1=集計,2=テンソル
    ReturnType:   pb.DataType_STRING,          // 関数戻り値=0=文字列,1=数値,2=Dual
    Params: []*pb.Parameter{
      &pb.Parameter {
        Name: "col1",                          // パラメータ名
        DataType: pb.DataType_STRING,          // パラメータタイプ=0=文字列,1=数値,2=Dual
      },
    },
  }
  
  // BigAdd
  var func1 = pb.FunctionDefinition {
    FunctionId:   1,                           // 関数ID
    Name:         "BigAdd",                    // 関数名
    FunctionType: pb.FunctionType_TENSOR,      // 関数タイプ=0=スカラー,1=集計,2=テンソル
    ReturnType:   pb.DataType_STRING,          // 関数戻り値=0=文字列,1=数値,2=Dual
    Params: []*pb.Parameter{
      &pb.Parameter {
        Name: "col1",                          // パラメータ名
        DataType: pb.DataType_STRING,          // パラメータタイプ=0=文字列,1=数値,2=Dual
      },
      &pb.Parameter {
        Name: "col2",                          // パラメータ名
        DataType: pb.DataType_STRING,          // パラメータタイプ=0=文字列,1=数値,2=Dual
      },
    },
  }
  
  var funcs = []*pb.FunctionDefinition {}
  funcs = append(funcs, &func0)
  funcs = append(funcs, &func1)
  capabilities.Functions = funcs
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
