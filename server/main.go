package main

import (
	"context"
	"log"
	"net"

	"strings"

	"google.golang.org/grpc"

	pb "ezcoinrobot/protos"
	"ezcoinrobot/server/funding"
	"ezcoinrobot/server/utils"

	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedEZCoinRobotServer
}

const (
	port                 = ":50052"
	supervisorConfigPath = "/etc/supervisor/conf.d/"
	configTemplate       = `[program:%name%_%currency%]
command=/home/john/SuperFundingBot/.venv/bin/python /home/john/bitfinex-funding-robot/create_funding_offers3.py %name% -s %currency%
autostart=true
autorestart=true
stderr_logfile=/var/log/ifund/%name%_%currency%.err.log
stdout_logfile=/var/log/ifund/%name%_%currency%.out.log
user=john
`
)

func (s *Server) CreateRobot(ctx context.Context, in *pb.RobotRequest) (*pb.CreateReply, error) {
	log.Printf("CreateRobot User: %s, Currency: %s", in.Name, in.Currency)

	if !funding.LegacyCreateSupervisorConfig(in.Name, in.Currency) {
		return &pb.CreateReply{Code: 1, Message: "create failed"}, nil
	}

	if err := utils.UpdateSupervisor(); err != nil {
		return &pb.CreateReply{Code: 1, Message: "update failed"}, err
	}

	return &pb.CreateReply{Message: "create success"}, nil
}

func (s *Server) RobotStatus(ctx context.Context, in *pb.RobotRequest) (*pb.StatusReply, error) {
	state, err := funding.LegacyRobotState(in.Name, in.Currency)
	if err != nil {
		return &pb.StatusReply{Code: 1, Message: err.Error()}, nil
	}
	return &pb.StatusReply{Code: 0, State: state}, nil
}

func (s *Server) StopRobot(ctx context.Context, in *pb.RobotRequest) (*pb.StatusReply, error) {
	serviceName := funding.LegacyRobotServiceName(in.Name, in.Currency)
	result, err := utils.DoAction(serviceName, "stop")
	if err != nil {
		return &pb.StatusReply{Code: 1, State: "", Message: ""}, err
	}

	log.Println(result)
	if strings.Contains(result, "ERROR") {
		return &pb.StatusReply{Code: 1, State: "", Message: result}, err
	}
	return &pb.StatusReply{Code: 0, State: "", Message: result}, nil
}

func (s *Server) StartRobot(ctx context.Context, in *pb.RobotRequest) (*pb.StatusReply, error) {
	serviceName := funding.LegacyRobotServiceName(in.Name, in.Currency)
	result, err := utils.DoAction(serviceName, "stop")
	if err != nil {
		return &pb.StatusReply{Code: 1, State: "", Message: ""}, err
	}
	log.Println(result)
	if strings.Contains(result, "ERROR") {
		return &pb.StatusReply{Code: 1, State: "", Message: result}, err
	}
	return &pb.StatusReply{Code: 0, State: "", Message: result}, nil
}

func (s *Server) RestartRobot(ctx context.Context, in *pb.RobotRequest) (*pb.StatusReply, error) {
	serviceName := funding.LegacyRobotServiceName(in.Name, in.Currency)
	result, err := utils.DoAction(serviceName, "restart")
	if err != nil {
		return &pb.StatusReply{Code: 1, State: "", Message: result}, err
	}

	log.Println(result)
	return &pb.StatusReply{Code: 0, State: "", Message: result}, nil
}

func (s *Server) MigrateRobot(ctx context.Context, in *pb.RobotMigrateRequest) (*pb.StatusReply, error) {
	if err := funding.LegacyReplaceRobotCurrent(in.Name, in.FromCurrency, in.ToCurrency); err != nil {
		return &pb.StatusReply{Code: 1, State: "", Message: err.Error()}, err
	}

	result, err := funding.RestartFundingRobot(in.Name, in.ToCurrency)
	if err != nil {
		return &pb.StatusReply{Code: 1, State: "", Message: result}, err
	}
	return &pb.StatusReply{Code: 0, State: "", Message: result}, nil
}

func (s *Server) CreateFundingRobot(ctx context.Context, in *pb.FundingRobotRequest) (*pb.CreateFundingReply, error) {
	log.Printf("CreateFundingRobot User: %s, Currency: %s", in.Name, in.Currency)
	if err := funding.CreateSupervisorFunding(in.Name, in.Currency); err != nil {
		return &pb.CreateFundingReply{Code: 1, Message: "Create funding robot failed"}, err
	}

	log.Println("execute supervisorctl update")
	if err := utils.UpdateSupervisor(); err != nil {
		return &pb.CreateFundingReply{Code: 2, Message: "Update supervisor failed"}, err
	}

	log.Println("execute supervisorctl update success")
	return &pb.CreateFundingReply{Message: "create success"}, nil
}

func (s *Server) FundingRobotStatus(ctx context.Context, in *pb.FundingRobotRequest) (*pb.FundingStatusReply, error) {
	state, err := funding.FundingRobotState(in.Name, in.Currency)
	if err != nil {
		return &pb.FundingStatusReply{Code: 1, Message: err.Error()}, nil
	}
	return &pb.FundingStatusReply{Code: 0, State: state}, nil
}

func (s *Server) RestartFundingRobot(ctx context.Context, in *pb.FundingRobotRequest) (*pb.FundingStatusReply, error) {
	result, err := funding.RestartFundingRobot(in.Name, in.Currency)
	if err != nil {
		return &pb.FundingStatusReply{Code: 1, State: "", Message: result}, err
	}

	log.Println(result)
	return &pb.FundingStatusReply{Code: 0, State: "", Message: result}, nil
}

func (s *Server) StopFundingRobot(ctx context.Context, in *pb.FundingRobotRequest) (*pb.FundingStatusReply, error) {
	result, err := funding.StopFundingRobot(in.Name, in.Currency)
	if err != nil {
		return &pb.FundingStatusReply{Code: 1, State: "", Message: ""}, err
	}
	log.Println(result)
	if strings.Contains(result, "ERROR") {
		return &pb.FundingStatusReply{Code: 1, State: "", Message: result}, err
	}
	return &pb.FundingStatusReply{Code: 0, State: "", Message: result}, nil
}

func (s *Server) StartFundingRobot(ctx context.Context, in *pb.FundingRobotRequest) (*pb.FundingStatusReply, error) {
	result, err := funding.StartFundingRobot(in.Name, in.Currency)
	if err != nil {
		return &pb.FundingStatusReply{Code: 1, State: "", Message: ""}, err
	}
	log.Println(result)
	if strings.Contains(result, "ERROR") {
		return &pb.FundingStatusReply{Code: 1, State: "", Message: result}, err
	}
	return &pb.FundingStatusReply{Code: 0, State: "", Message: result}, nil
}

func (s *Server) MigrateFundingRobot(ctx context.Context, in *pb.FundingRobotMigrateRequest) (*pb.FundingStatusReply, error) {
	if err := funding.MigrateFundingRobotServiceName(in.Name, in.FromCurrency, in.ToCurrency); err != nil {
		return &pb.FundingStatusReply{Code: 1, State: "", Message: err.Error()}, err
	}

	return &pb.FundingStatusReply{Code: 0, State: ""}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterEZCoinRobotServer(s, &Server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	log.Println("Server is running...")
}
