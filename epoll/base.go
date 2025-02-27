package epoll_server

//
//import (
//	"im-server/util"
//	"os"
//	"strconv"
//)
//
//func StartEpoll() {
//
//	//初始化reactor
//
//	epollM := NewEpollM()
//	defer epollM.Close()
//
//	epollIp := os.Getenv("EPOLL_IP")
//	epollPort, _ := strconv.Atoi(os.Getenv("EPOLL_PORT"))
//	//开启监听
//	err := epollM.Listen(epollIp, epollPort)
//	if err != nil {
//		panic(err)
//	}
//
//	//创建epoll
//	err = epollM.CreateEpoll()
//	if err != nil {
//		panic(err)
//	}
//
//	//初始化完成
//
//	//异步处理epoll
//	go handleEpollSocket(epollM)
//
//	//心跳检测
//	//go handleOvertimeSocket(epollM)
//
//	util.LogPrintln("epoll 服务启动:", epollIp, epollPort)
//	select {}
//}
//
//func handleEpollSocket(epollM *EpollM) {
//	err := epollM.HandlerEpoll()
//	panic(err) //这里的协程panic会影响主线程
//}
//
//func handleOvertimeSocket(epollM *EpollM) {
//	epollM.CloseOvertimeConn()
//}
