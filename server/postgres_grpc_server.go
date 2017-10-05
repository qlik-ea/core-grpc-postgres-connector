package main

import (
	"fmt"
	"log"
	"net"
	"github.com/qlik-ea/postgres-grpc-connector/qlik"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"os"
	"runtime/pprof"
	"flag"
	"time"
	"golang.org/x/net/context"

	"github.com/qlik-ea/postgres-grpc-connector/postgres"
)

const (
	port = ":50051"
)
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile `file`")

type server struct{
	postgresReaders map[string]*postgres.PostgresReader
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func (this *server) ExecuteGenericCommand(context context.Context, genericCommand *qlik.GenericCommand) (*qlik.GenericCommandResponse, error) {
	return &qlik.GenericCommandResponse{Data: "{}"}, nil
}

func (this *server) GetData(dataOptions *qlik.GetDataOptions, stream qlik.Connector_GetDataServer) error {

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		fmt.Println("Started cpu profiling...")
		defer pprof.StopCPUProfile()
	}

	var t0 = makeTimestamp()

	var connectionString = dataOptions.Connection.ConnectionString
	if dataOptions.Connection.User != "" {
		connectionString = connectionString + ";user=" + dataOptions.Connection.User
	}
	if dataOptions.Connection.Password != "" {
		connectionString = connectionString + ";password=" + dataOptions.Connection.Password
	}

	if this.postgresReaders[connectionString] == nil {
		fmt.Println("Starting connection pool");
		fmt.Println(connectionString);
		var err2 error
		this.postgresReaders[connectionString], err2 = postgres.NewPostgresReader(connectionString)
		if err2 != nil {
			return err2
		}
	} else {
		fmt.Println("Reusing connection pool")
	}
	var getDataErr = this.postgresReaders[connectionString].GetData(dataOptions, stream)

	var t1 = makeTimestamp()
	fmt.Println("Time", t1 - t0, "ms")
	return getDataErr
}

func main() {
	var err error

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	var srv = &server{ make(map[string]*postgres.PostgresReader)}
	qlik.RegisterConnectorServer(s, srv)
	// Register reflection service on gRPC server.
	reflection.Register(s)
	fmt.Println("Server started", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return;
	}



}
