package grpc_base

import (
	"google.golang.org/grpc"
	rpc_controller "im-server/grpc/controller"
	"im-server/grpc/proto_service"
	"im-server/util"
	"log"
	"net"
	"os"
)

func StartRpc() {

	rpcAddress := os.Getenv("RPC_ADDRESS")
	listen, err := net.Listen("tcp", rpcAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	// 服务注册
	proto_service.RegisterIMRpcServer(s, &rpc_controller.ImRpcController{})
	//这里服务注册重复了
	//proto_service.RegisterIMRpcServer(s, &rpc_controller.GroupRpcController{})
	//proto_service.RegisterIMRpcServer(s, &rpc_controller.UserGroupRpcController{})
	//proto_service.RegisterIMRpcServer(s, &rpc_controller.MessageRpcController{})

	util.LogPrintln("gRPC listen on " + rpcAddress)

	if err := s.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
