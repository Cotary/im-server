package epoll_server

import (
	"os"
	"strconv"
)

type MultiEpollReactor struct {
	mainReactor *epollReactor
	subReactor  []*epollReactor
	allReactor  []*epollReactor
}

func (multi *MultiEpollReactor) getSubReactor() *epollReactor {
	return multi.subReactor[0]
}

func (multi *MultiEpollReactor) run() {
	for _, v := range multi.allReactor {
		go eventLoop(v) //todo 这里留意下
	}
}

//创建多reactor
func createMultiReactor(subNum int) (*MultiEpollReactor, error) {
	multiEpollReactor := &MultiEpollReactor{}
	main, err := creatReactor(0, true, multiEpollReactor)
	if err != nil {
		return nil, err
	}
	var subReactor []*epollReactor
	for i := 1; i <= subNum; i++ {
		subTemp, err := creatReactor(i, false, multiEpollReactor)
		if err != nil {
			return nil, err
		}
		subReactor = append(subReactor, subTemp)

	}
	var allReactor []*epollReactor
	allReactor = append(allReactor, main)
	allReactor = append(allReactor, subReactor...)

	multiEpollReactor.mainReactor = main
	multiEpollReactor.subReactor = subReactor
	multiEpollReactor.allReactor = allReactor

	return multiEpollReactor, nil
}

func ReactorRun() {
	//创建几个epoll_reactor
	multi, err := createMultiReactor(1)
	if err != nil {
		panic("reactor创建失败")
	}
	//创建监听套接字,并且放入main
	epollIp := os.Getenv("EPOLL_IP")
	epollPort, _ := strconv.Atoi(os.Getenv("EPOLL_PORT"))
	err = multi.mainReactor.Listen(epollIp, epollPort)
	if err != nil {
		panic(err.Error())
	}
	//开启监听,在监听里面处理事务
	multi.run()

	println("启动")
	select {}

}
