package epoll_server

//
//import (
//	"encoding/json"
//	"fmt"
//	"im-server/model"
//	errorServer "im-server/server/error"
//	"im-server/util"
//
//	"syscall"
//	"time"
//)
//
//type ServerConn struct {
//	epollM    *EpollM
//	fd        int
//	lastTime  int
//	userInfo  model.User //im用户id    校验通过，是正常用户
//	websocket *websocketClient
//}
//
////读取数据
//func (s *ServerConn) handle() {
//	content := make([]byte, 1024)
//	if s.websocket == nil {
//
//		//通过系统调用,读取数据,n是读到的长度
//		n, err := syscall.Read(s.fd, content)
//		if n == 0 || err != nil {
//			//如果n=0，那这个链接就是客户端异常关闭了
//			s.epollM.CloseConn(s.fd)
//			return
//		}
//		//判断是不是websocket ,有没有连接，没有就升级并且连接
//		if isWebsocketConnect(s, content) {
//			upgrade(s, content)
//			return
//		}
//	} else {
//		util.LogPrintln("ReadIframe来数据了")
//
//		var frameType byte
//		var err error
//		frameType, content, err = ReadIframe(s)
//		if err!=nil {
//			util.LogPrintln("ReadIframe   读数据失败")
//			return
//		}
//		if frameType == TextMessage {
//			//WriteIframe(s, content, TextMessage)
//		} else if frameType == PingMessage {
//			WriteIframe(s, []byte{}, PongMessage)
//		}
//	}
//	//更新最后登录时间
//	s.lastTime = int(time.Now().Unix())
//
//	//读取消息
//
//	//预处理数据 必须符合requestMessage 格式
//	//content[:] 注意这个地方要这么写，不然解析不出来,末尾会多个结束符
//	requestMessage, requestMessageErr := PreHandleMessage(content)
//	if requestMessageErr != nil {
//		s.SendErrorMsg(requestMessageErr.Error())
//		return
//	}
//
//	//上线操作
//	if s.userInfo == (model.User{}) {
//		userInfo := UserIMOnline(requestMessage)
//		s.userInfo = userInfo
//
//		util.LogPrintln(s.userInfo.Name, " 上线成功!")
//		//上线成功后发送离线消息
//		offlineErr := UserIMSendOfflineMessage(requestMessage, s)
//		if offlineErr != nil {
//			util.LogPrintln("发送离线消息失败：", offlineErr.Error())
//		}
//	} else {
//		//已在线，处理心跳
//		HandleHeartbeat(requestMessage)
//	}
//
//	//读取消息
//	if s.userInfo != (model.User{}) && requestMessage.Type == NormalMsg {
//
//		//发送消息socket
//		isSend := false
//		ToConn := s.epollM.GetUserConn(requestMessage.ToProjectId, requestMessage.ToProjectUid)
//		if ToConn != nil {
//			util.LogPrintln(s.userInfo.Name, " 给 ", ToConn.userInfo.Name, " 发送消息：", requestMessage.Data)
//
//			ToConn.SendMsg(requestMessage)
//			isSend = true
//		}
//
//		//存下消息
//		err := UserIMSaveMessage(requestMessage, isSend)
//		if err != nil {
//			util.LogPrintln("消息存储失败")
//		}
//
//	}
//
//	//回执消息
//	s.SendReceiptMsg(requestMessage)
//
//}
//
//func (s *ServerConn) SendMsg(response IMMessage) {
//	responseJson, _ := json.Marshal(response)
//	s.Send(responseJson)
//}
//
//func (s *ServerConn) SendErrorMsg(msg string) {
//	responseStruct := IMMessage{
//		Time: int(time.Now().Unix()),
//		Type: ErrorMsg,
//		Data: msg,
//	}
//	responseJson, _ := json.Marshal(responseStruct)
//	s.Send(responseJson)
//}
//
//func (s *ServerConn) SendReceiptMsg(request IMMessage) {
//	responseStruct := IMMessage{
//		Time: int(time.Now().Unix()),
//		Type: ReceiptMsg,
//		Data: request.MessageId, //回执消息，雪花算法id
//	}
//	responseJson, _ := json.Marshal(responseStruct)
//	s.Send(responseJson)
//}
//
////向这个链接中写数据
//func (s *ServerConn) Send(data []byte) {
//	if s.websocket != nil {
//		WriteIframe(s, data, TextMessage)
//	} else {
//		s.Write(data)
//	}
//}
//
////向这个链接中读数据
//func (s *ServerConn) Read(data []byte) (int, error) {
//
//	n, err := syscall.Read(s.fd, data)
//	if err != nil {
//		//s.epollM.CloseConn(s.fd)
//		fmt.Print(errorServer.WrapError(errorServer.E("Read error"), ""))
//		util.LogPrintln(s.fd, "Read error:", err.Error())
//	}
//	return n, err
//}
//
////向这个链接中写数据
//func (s *ServerConn) Write(data []byte) {
//	_, err := syscall.Write(s.fd, data)
//	if err != nil {
//		util.LogPrintln(s.fd, "Write error:", err.Error())
//	}
//}
//
////关闭这个链接
//func (s *ServerConn) Close() error {
//	err := syscall.Close(s.fd)
//	return err
//}
