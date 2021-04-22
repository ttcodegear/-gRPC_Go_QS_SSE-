package main

import (
  "log"
  "net"
  "math"
  "io"
  "context"
  "github.com/golang/protobuf/proto"
  "google.golang.org/grpc"
  "google.golang.org/grpc/metadata"
  "google.golang.org/grpc/status"
  "google.golang.org/grpc/codes"
  pb "sse/pb/sse"
  . "github.com/ahmetb/go-linq/v3"
  mapset "github.com/deckarep/golang-set"
)

type ConnectorServer struct {
  pb.UnimplementedConnectorServer
}

// Function Name | Function Type  | Argument     | TypeReturn Type
// ScriptEval    | Scalar, Tensor | Numeric      | Numeric
// ScriptEvalEx  | Scalar, Tensor | Dual(N or S) | Numeric
func (s *ConnectorServer) ScriptEval(header pb.ScriptRequestHeader, stream pb.Connector_ExecuteFunctionServer) error {
  log.Printf("script=" + header.Script)
  // パラメータがあるか否かをチェック
  if( len(header.Params) > 0 ) {
    return nil
  } else {
    var result float64 = math.NaN()
    for {
      _, err := stream.Recv()
      if err == io.EOF {
        
        result = 12.34
        
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
    }
  }
}

// Function Name   | Function Type | Argument     | TypeReturn Type
// ScriptAggrStr   | Aggregation   | String       | String
// ScriptAggrExStr | Aggregation   | Dual(N or S) | String
func (s *ConnectorServer) ScriptAggrStr(header pb.ScriptRequestHeader, stream pb.Connector_ExecuteFunctionServer) error {
  log.Printf("script=" + header.Script)
  // パラメータがあるか否かをチェック
  if( len(header.Params) > 0 ) {
    return nil
  } else {
    return nil
  }
}

func (s *ConnectorServer) GetFunctionName(header pb.ScriptRequestHeader) string {
  func_type := header.FunctionType
  arg_types := []interface{}{ }
  for _, param := range header.Params {
    arg_types = append(arg_types, param.DataType)
  }
  ret_type := header.ReturnType
/*
  if func_type == pb.FunctionType_SCALAR || func_type == pb.FunctionType_TENSOR {
    log.Printf("func_type SCALAR TENSOR")
  } else if func_type == pb.FunctionType_AGGREGATION {
    log.Printf("func_type AGGREGATION")
  }

  if len(arg_types) == 0 {
    log.Printf("arg_type Empty")
  } else if From(arg_types).All(func(i interface{}) bool{ return i.(pb.DataType) == pb.DataType_NUMERIC }) {
    log.Printf("arg_type NUMERIC")
  } else if From(arg_types).All(func(i interface{}) bool{ return i.(pb.DataType) == pb.DataType_STRING }) {
    log.Printf("arg_type STRING")
  } else if (mapset.NewSetFromSlice(arg_types)).Cardinality() >= 2 ||
            From(arg_types).All(func(i interface{}) bool{ return i.(pb.DataType) == pb.DataType_DUAL }) {
    log.Printf("arg_type DUAL")
  }

  if ret_type == pb.DataType_NUMERIC {
    log.Printf("ret_type NUMERIC")
  } else if ret_type == pb.DataType_STRING {
    log.Printf("ret_type STRING")
  }
*/
  if func_type == pb.FunctionType_SCALAR || func_type == pb.FunctionType_TENSOR {
    if len(arg_types) == 0 ||
       From(arg_types).All(func(i interface{}) bool{ return i.(pb.DataType) == pb.DataType_NUMERIC }) {
      if ret_type == pb.DataType_NUMERIC {
        return "ScriptEval"
      }
    }
  }
  
  if func_type == pb.FunctionType_SCALAR || func_type == pb.FunctionType_TENSOR {
    if (mapset.NewSetFromSlice(arg_types)).Cardinality() >= 2 ||
       From(arg_types).All(func(i interface{}) bool{ return i.(pb.DataType) == pb.DataType_DUAL }) {
      if ret_type == pb.DataType_NUMERIC {
        return "ScriptEvalEx"
      }
    }
  }
  
  if func_type == pb.FunctionType_AGGREGATION {
    if len(arg_types) == 0 ||
       From(arg_types).All(func(i interface{}) bool{ return i.(pb.DataType) == pb.DataType_STRING }) {
      if ret_type == pb.DataType_STRING {
        return "ScriptAggrStr"
      }
    }
  }
  
  if func_type == pb.FunctionType_AGGREGATION {
    if (mapset.NewSetFromSlice(arg_types)).Cardinality() >= 2 ||
       From(arg_types).All(func(i interface{}) bool{ return i.(pb.DataType) == pb.DataType_DUAL }) {
      if ret_type == pb.DataType_STRING {
        return "ScriptAggrExStr"
      }
    }
  }
  
  return "Unsupported Function Name"
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
      func_name := s.GetFunctionName(header)
      if func_name == "ScriptEval" || func_name == "ScriptEvalEx" {
        return s.ScriptEval(header, stream)
      }
      if func_name == "ScriptAggrStr" || func_name == "ScriptAggrExStr" {
        return s.ScriptAggrStr(header, stream)
      }
    }
  }
  return status.Errorf(codes.Unimplemented, "Method not implemented!")
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
