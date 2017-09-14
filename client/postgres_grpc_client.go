package main

import (
	"github.com/qlik-ea/postgres-grpc-connector/qlik"
	"google.golang.org/grpc"
	"github.com/golang/protobuf/proto"
	"context"
	"fmt"
	"time"
)

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func main() {

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	client := qlik.NewConnectorClient(conn)
	var getDataOptions = &qlik.GetDataOptions{}
	getDataOptions.Connection = &qlik.ConnectionInfo{"host=selun.gwe.qliktech.com;user=testuser;password=testuser;database=test", "",""}
	getDataOptions.Parameters = &qlik.DataInfo{"select * from airports", ""}
	var t0 = makeTimestamp()

	var stream, err2 = client.GetData(context.Background(), getDataOptions)
	fmt.Println(err2)
	var header, _ = stream.Header()
	var t = header["x-qlik-getdata-bin"]
	var t2 = t[0]
	var dataResponse = qlik.GetDataResponse{FieldInfo: make([]*qlik.FieldInfo, 100), TableName: "x"}
	proto.Unmarshal([]byte(t2), &dataResponse)
	fmt.Println("a", t)

	if err2 != nil {
		fmt.Println(err)
	}
	var bundle, receiveError = stream.Recv()
	if bundle.Rows != nil {
		fmt.Println("cells by row")
	} else {
		fmt.Println("cells by column")
	}

	var totalCount int
	for receiveError == nil {
		if bundle.Cols != nil {
			var stringsLen = len(bundle.Cols[0].Strings)
			var numbersLen = len(bundle.Cols[0].Numbers)
			if stringsLen > 0 {
				totalCount += stringsLen
			} else {
				totalCount += numbersLen
			}
		} else {
			totalCount += len(bundle.Rows)
		}

		bundle, receiveError = stream.Recv()
	}
	var t1 = makeTimestamp()
	fmt.Println("Total rows", totalCount)
	fmt.Println("Time", t1-t0, "ms")

}
