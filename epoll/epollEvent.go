package epoll_server

import (
	"encoding/json"
	"fmt"
	"im-server/model"
	"syscall"
	"time"
)

type event struct {
	fd           int
	epollReactor *epollReactor
	readBuff     []byte //todo 这个buff要清理，go应该不用
	writeBuff    []byte
	//todo 回调

	lastTime  int
	userInfo  model.User //im用户id    校验通过，是正常用户
	websocket *websocketClient
}

func (event *event) handleHup() {
	fmt.Println("socket hup")
	event.epollReactor.closeEvent(event)
}
func (event *event) handleErr() {
	//n, err := syscall.GetsockoptInt(event.fd, syscall.SOL_SOCKET, syscall.SO_ERROR)
	//syscall.Errno.Error(n)
	//if n < 0 {
	//
	//}
	fmt.Println("socket err")
	event.epollReactor.closeEvent(event)
}

func (event *event) handleIn() {
	if event.epollReactor.isMain {
		//accept
		fmt.Println("accept")
		err := event.epollReactor.accept()
		if err != nil {
			fmt.Println("accept err", err.Error())
		}
	} else {
		fmt.Println("read")
		//read  decode process encode send
		event.process()
	}
}

func (event *event) handleOut() {
	//把缓冲区的数据写进去
	if len(event.writeBuff) > 0 {
		event.writeSocket(event.writeBuff)
	}
}

func (e *event) process() error {
	//read
	e.readSocket()
	content := e.readBuff
	e.readBuff = []byte{}
	fmt.Println("orginal:", content)

	frame := decodeData(e, content)
	if len(frame) > 0 {
		fmt.Println("context:", string(frame))

		//process
		HandleBusiness(e, frame)
	}

	//send
	return nil
}

func (e *event) readSocket() int {
	//read all
	num := 0
	for {
		buff := make([]byte, 1024)
		n, err := syscall.Read(e.fd, buff)
		fmt.Println(n, err, err == syscall.EWOULDBLOCK)
		if n == 0 {
			//关闭了socket
			e.epollReactor.closeEvent(e)
			break
		} else if n < 0 {
			if err == syscall.EINTR {
				//信号打断了
				continue
			} else if err == syscall.EWOULDBLOCK {
				//读缓存区为空，没得读了，就返回了
				break
			} else {
				//出错了，关闭了socket
				e.epollReactor.closeEvent(e)
				break
			}
		} else {
			//读到数据了，写入缓冲区
			e.readBuff = append(e.readBuff, buff[:n]...)
		}
		num += n
	}
	return num
}
func (e *event) writeSocket(data []byte) {
	write_num := e._writeSocket(data)
	if write_num == 0 && len(data) > 0 {
		//清空缓冲区。todo 这里还要考虑剩下的数据吧，应该把剩下的数据放入缓冲区？
		e.writeBuff = []byte{}
		e.writeBuff = append(e.writeBuff, data...)
		e.epollReactor.editEvent(e.fd, syscall.EPOLLIN|syscall.EPOLLOUT)
		//加了可写的监听，后续怎么办
	} else if write_num > 0 {
		//还原
		e.epollReactor.editEvent(e.fd, syscall.EPOLLIN)
	}
}

func (e *event) _writeSocket(data []byte) int {
	for {
		n, err := syscall.Write(e.fd, data)
		fmt.Println("write", n, err, err == syscall.EWOULDBLOCK)
		if n < 0 {
			if err == syscall.EINTR {
				//信号打断了
				continue
			} else if err == syscall.EWOULDBLOCK {
				//写缓冲区满了
				break
			} else {
				//出错了，关闭了socket
				e.epollReactor.closeEvent(e)
				break
			}

		}
		return n
	}
	return 0
}

func newEvent(epollReactor *epollReactor, fd int) *event {
	return &event{
		fd:           fd,
		epollReactor: epollReactor,
	}
}

func creatReactor(code int, isMain bool, multiEpollReactor *MultiEpollReactor) (*epollReactor, error) {
	reactor := &epollReactor{}

	//系统调用EpollCreate
	fd, err := syscall.EpollCreate(1)
	if err != nil {
		return nil, err
	}
	reactor.epollFd = fd
	reactor.isMain = isMain
	reactor.code = code
	reactor.multiEpollReactor = multiEpollReactor
	return reactor, nil

}

// 封装的im相关的方法
func (e *event) SendReceiptMsg(request IMMessage) {
	responseStruct := IMMessage{
		Time: int(time.Now().Unix()),
		Type: ReceiptMsg,
		Data: request.MessageId, //回执消息，雪花算法id
	}
	responseJson, _ := json.Marshal(responseStruct)
	e.Send(responseJson)
}

// 向这个链接中写数据
func (e *event) Send(data []byte) {
	if e.websocket != nil {
		data = GetIframeByte(e, data, TextMessage)
	}
	e.writeSocket(data)
}

func (e *event) SendMsg(response IMMessage) {
	responseJson, _ := json.Marshal(response)
	e.Send(responseJson)
}
func (e *event) SendErrorMsg(msg string) {
	responseStruct := IMMessage{
		Time: int(time.Now().Unix()),
		Type: ErrorMsg,
		Data: msg,
	}
	responseJson, _ := json.Marshal(responseStruct)
	e.Send(responseJson)
}

// 获取符合条件的用户
func (e *event) GetUserConn(projectId int, projectUid int) (res *event) {

	e.epollReactor.events.Range(func(k, v interface{}) bool {
		conn := v.(*event)
		if conn.userInfo.ProjectId == projectId && conn.userInfo.ProjectUid == projectUid {
			res = conn
			return false
		}
		return true
	})
	return res

}

//设置监听socket，
//eopll池子，一个拿来监听，其他拿来处理
//把监听socket放入监听epoll，epoll检测,有数据就accept，放入某一个的监听队列，
