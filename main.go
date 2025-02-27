package main

import (
	"bytes"
	"fmt"
	"github.com/joho/godotenv"
	"im-server/cache"
	"im-server/epoll"
	"im-server/grpc"
	"im-server/model"
	"im-server/util"
	"log"
	"os"
	"runtime"
	// "syscall"
)

func init() {
	// 获取日志文件句柄
	// 以 只写入文件|没有时创建|文件尾部追加 的形式打开这个文件
	logFile, err := os.OpenFile(`./log/test.log`, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	// 设置存储位置
	log.SetOutput(logFile)
}

func main() {
	defer logPanic()
	RewriteStderrFile()

	//从本地读取环境变量
	godotenv.Load()
	model.Init()
	//开启redis
	cache.Redis()

	//grpc
	go grpc_base.StartRpc()

	//epoll
	go epoll_server.ReactorRun()

	//for i:=1;i<5;i++{
	//    go func() {
	//
	//        time.Sleep(10)
	//        fmt.Println(runtime.NumGoroutine())
	//        select {
	//
	//        }
	//    }()
	//}
	//fmt.Println("over")
	select {}
}

var stdErrFileHandler *os.File

const stdErrFile = "./log/test2.log"

func RewriteStderrFile() error {
	//if runtime.GOOS == "windows" {
	//    return nil
	//}
	//
	//file, err := os.OpenFile(stdErrFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	//if err != nil {
	//    fmt.Println(err)
	//    return err
	//}
	//stdErrFileHandler = file //把文件句柄保存到全局变量，避免被GC回收
	//
	//if err = syscall.Dup2(int(file.Fd()), int(os.Stderr.Fd())); err != nil {
	//    fmt.Println(err)
	//    return err
	//}
	//// 内存回收前关闭文件描述符
	//runtime.SetFinalizer(stdErrFileHandler, func(fd *os.File) {
	//    fd.Close()
	//})
	//
	return nil
}

func logPanic() {
	if info := recover(); info != nil {

		util.LogPrintln("PANIC_TRACK", PanicTrace(info))
	}
}

func PanicTrace(err interface{}) string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%v\n", err)
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
	}
	return buf.String()
}
