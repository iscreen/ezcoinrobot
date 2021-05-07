package main

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	pb "ezcoinrobot/protos"
)

const (
	defaultName    = "dean.lin"
	defaultCurreny = "fUSD"
)

func main() {
	EZCoinServer := os.Getenv("EZCOIN_SERVER")
	log.Printf("EZCoin server %s \n", EZCoinServer)

	// Set up a connection to the server.
	conn, err := grpc.Dial(EZCoinServer, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewEZCoinRobotClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.CreateFundingRobot(ctx, &pb.FundingRobotRequest{Name: name, Currency: defaultCurreny})
	if err != nil {
		log.Fatalf("could not create robot: %v", err)
	}
	log.Printf("CreateRobot: code: %d, message: %s", r.Code, r.Message)

	x, err := c.MigrateFundingRobot(ctx, &pb.FundingRobotMigrateRequest{Name: name, FromCurrency: defaultCurreny, ToCurrency: "fUST"})
	if err != nil {
		log.Fatalf("could not create robot: %v", err)
	}
	log.Printf("CreateRobot: code: %d, message: %s", x.Code, x.Message)
}
