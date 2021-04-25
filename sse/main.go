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
)

type ConnectorServer struct {
  pb.UnimplementedConnectorServer
}

func (s *ConnectorServer) SumOfColumn(stream pb.Connector_ExecuteFunctionServer) error {
  log.Printf("SumOfColumn")
  var params []float64
  for {
    in, err := stream.Recv()
    if err == io.EOF {
      var result float64 = 0.0
      for _, num := range params {
        result += num                               // Col1 + Col1 + ...
      }
      dual := pb.Dual{ NumData: result }
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
      params = append(params, row.Duals[0].NumData) // row=[Col1]
    }
  }
}

func (s *ConnectorServer) SumOfRows(stream pb.Connector_ExecuteFunctionServer) error {
  log.Printf("SumOfRows")
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
      param1 := row.Duals[0].NumData // row=[Col1,Col2]
      param2 := row.Duals[1].NumData
      result := param1 + param2      // Col1 + Col2
      dual := pb.Dual{ NumData: result }
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
    return s.SumOfColumn(stream)
  } else if(func_id == 1) {
    return s.SumOfRows(stream)
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
  
  // SumOfColumn
  var func0 = pb.FunctionDefinition {
    FunctionId:   0,                           // 関数ID
    Name:         "SumOfColumn",               // 関数名
    FunctionType: pb.FunctionType_AGGREGATION, // 関数タイプ=0=スカラー,1=集計,2=テンソル
    ReturnType:   pb.DataType_NUMERIC,         // 関数戻り値=0=文字列,1=数値,2=Dual
    Params: []*pb.Parameter{
      &pb.Parameter {
        Name: "col1",                          // パラメータ名
        DataType: pb.DataType_NUMERIC,         // パラメータタイプ=0=文字列,1=数値,2=Dual
      },
    },
  }
  
  // SumOfRows
  var func1 = pb.FunctionDefinition {
    FunctionId:   1,                           // 関数ID
    Name:         "SumOfRows",                 // 関数名
    FunctionType: pb.FunctionType_TENSOR,      // 関数タイプ=0=スカラー,1=集計,2=テンソル
    ReturnType:   pb.DataType_NUMERIC,         // 関数戻り値=0=文字列,1=数値,2=Dual
    Params: []*pb.Parameter{
      &pb.Parameter {
        Name: "col1",                          // パラメータ名
        DataType: pb.DataType_NUMERIC,         // パラメータタイプ=0=文字列,1=数値,2=Dual
      },
      &pb.Parameter {
        Name: "col2",                          // パラメータ名
        DataType: pb.DataType_NUMERIC,         // パラメータタイプ=0=文字列,1=数値,2=Dual
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
