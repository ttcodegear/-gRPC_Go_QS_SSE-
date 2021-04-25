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
  . "github.com/ahmetb/go-linq/v3"
  mapset "github.com/deckarep/golang-set"

  "math"
  "errors"
  "fmt"
  "os"
  "io/ioutil"
  "strings"
  "os/exec"
  "bytes"
  "runtime"
  "encoding/gob"
  "reflect"
  "strconv"
)

type ConnectorServer struct {
  pb.UnimplementedConnectorServer
}

type SSEArgs struct {
  AllArgs [][]interface{}
}

type SSERetVals struct {
  RetVals []interface{}
}

type ParamTuple struct {
  dtype pb.DataType
  dual  *pb.Dual
}

func ExecGoRun(script_txt string, template_name string, all_args *SSEArgs, all_results *SSERetVals) error {
  tmp_dir := "tmp"
  if _, err := os.Stat(tmp_dir); os.IsNotExist(err) {
    os.Mkdir(tmp_dir, 0777)
  }
  
  gobfile, err := ioutil.TempFile(tmp_dir, "temp.*.gob")
  if err != nil {
    log.Printf("%v", err)
    return errors.New("Go Run Error: Failed to create .gob file")
  }
  encoder := gob.NewEncoder(gobfile)
  if err := encoder.Encode(all_args); err != nil {
    log.Printf("%v", err)
    return errors.New("Go Run Error: Failed to encode arguments(.gob file)")
  }
  gobfile.Close()
  
  tempcontent, err := ioutil.ReadFile(template_name)
  if err != nil {
    log.Printf("%v", err)
    return errors.New("Go Run Error: Failed to open template .go file")
  }
  gofile, err := ioutil.TempFile(tmp_dir, "temp.*.go")
  if err != nil {
    log.Printf("%v", err)
    return errors.New("Go Run Error: Failed to create .go file")
  }
  content := strings.Replace(string(tempcontent), "$$$", script_txt, 1)
  ioutil.WriteFile(gofile.Name(), []byte(content), os.ModePerm)
  gofile.Close()
  
  var stdout bytes.Buffer
  var stderr bytes.Buffer
  var cmd *exec.Cmd
  goRun := fmt.Sprintf("go run %s %s", gofile.Name(), gobfile.Name())
  if runtime.GOOS == "windows" {
    cmd = exec.Command("cmd", "/C", goRun)
  } else {
    cmd = exec.Command("bash", "-c", goRun)
  }
  cmd.Stdout = &stdout
  cmd.Stderr = &stderr
  if err := cmd.Run(); err != nil {
    log.Printf("%v", err)
  }
  err_result := stderr.String()
  fmt.Print(err_result)
  std_result := stdout.String()
  fmt.Print(std_result)
  
  retfile, err := os.Open(gobfile.Name())
  if err != nil {
    log.Printf("%v", err)
    return errors.New("Go Run Error: Failed to open .gob file for results")
  }
  decoder := gob.NewDecoder(retfile)
  if err := decoder.Decode(all_results); err != nil {
    log.Printf("%v", err)
    return errors.New("Go Run Error: Failed to decode results(.gob file)")
  }
  retfile.Close()

  if err := os.Remove(retfile.Name()); err != nil {
    log.Printf("%v", err)
  }
  if err := os.Remove(gofile.Name()); err != nil {
    log.Printf("%v", err)
  }
  
  return nil
}

// Function Name | Function Type  | Argument     | TypeReturn Type
// ScriptEval    | Scalar, Tensor | Numeric      | Numeric
// ScriptEvalEx  | Scalar, Tensor | Dual(N or S) | Numeric
func (s *ConnectorServer) ScriptEval(header pb.ScriptRequestHeader, stream pb.Connector_ExecuteFunctionServer) error {
  log.Printf("script=" + header.Script)
  // パラメータがあるか否かをチェック
  if( len(header.Params) > 0 ) {
    for {
      in, err := stream.Recv()
      if err == io.EOF {
        return nil
      }
      if err != nil {
        return err
      }
      var all_args SSEArgs = SSEArgs{}
      for _, row := range in.Rows {
        script_args := []interface{}{}
        zip := make([]ParamTuple, len(header.Params), len(header.Params))
        for i, param := range header.Params {
          zip[i] = ParamTuple{ param.DataType, row.Duals[i] }
        }
        for _, elm := range zip {
          if  elm.dtype == pb.DataType_NUMERIC || elm.dtype == pb.DataType_DUAL {
            script_args = append(script_args, elm.dual.NumData)
          } else {
            script_args = append(script_args, elm.dual.StrData)
          }
        }
        log.Printf("args=%v", script_args)
        all_args.AllArgs = append(all_args.AllArgs, script_args)
      }
      var all_results *SSERetVals = &SSERetVals{}
      if err := ExecGoRun(header.Script, "ScriptEval.txt", &all_args, all_results); err != nil {
        return err
      }
      var response_rows pb.BundledRows
      for _, d := range all_results.RetVals {
        var result float64 = math.NaN()
        kind := reflect.TypeOf(d).Kind()
        val := reflect.ValueOf(d)
        if kind == reflect.Float64 {
          result = val.Float()
        } else if kind == reflect.String {
          f, err := strconv.ParseFloat(val.String(), 64)
          if err != nil {
            f = math.NaN()
          }
          result = f
        } else if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 {
          result = float64(val.Int())
        } else if kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 {
          result = float64(val.Uint())
        } else {
          result = math.NaN()
        }
        dual := pb.Dual{ NumData: result }
        r := pb.Row{ Duals: []*pb.Dual{ &dual } }
        response_rows.Rows = append(response_rows.Rows, &r)
      }
      if err := stream.Send(&response_rows); err != nil {
        return err
      }
    }
  } else {
    for {
      _, err := stream.Recv()
      if err == io.EOF {
        var all_args SSEArgs = SSEArgs{}
        var all_results *SSERetVals = &SSERetVals{}
        if err := ExecGoRun(header.Script, "ScriptEval.txt", &all_args, all_results); err != nil {
          return err
        }
        var result float64 = math.NaN()
        for _, d := range all_results.RetVals {
          kind := reflect.TypeOf(d).Kind()
          val := reflect.ValueOf(d)
          if kind == reflect.Float64 {
            result = val.Float()
          } else if kind == reflect.String {
            f, err := strconv.ParseFloat(val.String(), 64)
            if err != nil {
              f = math.NaN()
            }
            result = f
          } else if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 {
            result = float64(val.Int())
          } else if kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 {
            result = float64(val.Uint())
          } else {
            result = math.NaN()
          }
        }
        var reply pb.BundledRows
        dual := pb.Dual{ NumData: result }
        r := pb.Row{ Duals: []*pb.Dual{ &dual } }
        reply.Rows = append(reply.Rows, &r)
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
    var all_args SSEArgs = SSEArgs{}
    for {
      in, err := stream.Recv()
      if err == io.EOF {
        var all_results *SSERetVals = &SSERetVals{}
        if err := ExecGoRun(header.Script, "ScriptAggrStr.txt", &all_args, all_results); err != nil {
          return err
        }
        var response_rows pb.BundledRows
        for _, d := range all_results.RetVals {
          var result string = ""
          kind := reflect.TypeOf(d).Kind()
          val := reflect.ValueOf(d)
          if kind == reflect.String {
            result = val.String()
          } else if kind == reflect.Float64 {
            result = strconv.FormatFloat(val.Float(), 'f', -1, 64)
          } else if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 {
            result = strconv.FormatInt(val.Int(), 10)
          } else if kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 {
            result = strconv.FormatUint(val.Uint(), 10)
          } else {
            result = ""
          }
          dual := pb.Dual{ StrData: result }
          r := pb.Row{ Duals: []*pb.Dual{ &dual } }
          response_rows.Rows = append(response_rows.Rows, &r)
        }
        if err := stream.Send(&response_rows); err != nil {
          return err
        }
        return nil
      }
      if err != nil {
        return err
      }
      for _, row := range in.Rows {
        script_args := []interface{}{}
        zip := make([]ParamTuple, len(header.Params), len(header.Params))
        for i, param := range header.Params {
          zip[i] = ParamTuple{ param.DataType, row.Duals[i] }
        }
        for _, elm := range zip {
          if  elm.dtype == pb.DataType_STRING || elm.dtype == pb.DataType_DUAL {
            script_args = append(script_args, elm.dual.StrData)
          } else {
            script_args = append(script_args, elm.dual.NumData)
          }
        }
        log.Printf("args=%v", script_args)
        all_args.AllArgs = append(all_args.AllArgs, script_args)
      }
    }
  } else {
    for {
      _, err := stream.Recv()
      if err == io.EOF {
        var all_args SSEArgs = SSEArgs{}
        var all_results *SSERetVals = &SSERetVals{}
        if err := ExecGoRun(header.Script, "ScriptAggrStr.txt", &all_args, all_results); err != nil {
          return err
        }
        var result string = ""
        for _, d := range all_results.RetVals {
          kind := reflect.TypeOf(d).Kind()
          val := reflect.ValueOf(d)
          if kind == reflect.String {
            result = val.String()
          } else if kind == reflect.Float64 {
            result = strconv.FormatFloat(val.Float(), 'f', -1, 64)
          } else if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 {
            result = strconv.FormatInt(val.Int(), 10)
          } else if kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 {
            result = strconv.FormatUint(val.Uint(), 10)
          } else {
            result = ""
          }
        }
        var reply pb.BundledRows
        dual := pb.Dual{ StrData: result }
        r := pb.Row{ Duals: []*pb.Dual{ &dual } }
        reply.Rows = append(reply.Rows, &r)
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
