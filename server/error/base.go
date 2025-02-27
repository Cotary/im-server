package errorServer

import (
	"errors"
	"fmt"
	"runtime"
)

func E(str string) error {
	err := errors.New(str)
	//util.LogPrintln(err)  注意这里造成了循环引用包
	return err
}

const defaultErrMsg = "Error occur:"

/**
 * 用特定信息新创建一个 error
 */
func NewErrorf(format string, a ...interface{}) error {
	pc, _, line, ok := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	if !ok {
		return errors.New("WrapError 方法获取堆栈失败")
	}

	errMsg := fmt.Sprintf("error occur, cause: %s \n\tat %s:%d", fmt.Sprintf(format, a...), f.Name(), line)
	return errors.New(errMsg)
}

/**
 * 用特定信息包装一个 error，使其包含代码堆栈信息
 * 如果 err 为空则返回空
 */
func WrapError(err error, wrapMsg string) error {
	pc, _, line, ok := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	if !ok {
		return errors.New("WrapError 方法获取堆栈失败")
	}

	var wrapErr error = nil
	if err != nil {
		if wrapMsg == "" {
			wrapMsg = defaultErrMsg
		}
		errMsg := fmt.Sprintf("%s \n\tat %s:%d\nCause by: %s\n", wrapMsg, f.Name(), line, err.Error())
		wrapErr = errors.New(errMsg)
	}
	return wrapErr
}

/**
 * 用特定信息包装一个 error，使其包含代码堆栈信息
 * 如果 err 为空则返回空
 */
func WrapErrorf(err error, wrapMsgFmt string, a ...interface{}) error {
	wrapMsg := ""
	if wrapMsgFmt == "" {
		wrapMsg = defaultErrMsg
	} else {
		wrapMsg = fmt.Sprintf(wrapMsgFmt, a...)
	}

	pc, _, line, ok := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	if !ok {
		return errors.New("WrapError 方法获取堆栈失败")
	}

	var wrapErr error = nil
	if err != nil {
		if wrapMsg == "" {
			wrapMsg = defaultErrMsg
		}
		errMsg := fmt.Sprintf("%s \n\tat %s:%d\nCause by: %s\n", wrapMsg, f.Name(), line, err.Error())
		wrapErr = errors.New(errMsg)
	}
	return wrapErr
}

/**
 * 获得错误描述的同时携带上代码堆栈信息
 */
func GetErrorStack(err error, preStr string) string {
	pc, _, line, ok := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	if !ok {
		return "GetErrorStack 方法获取堆栈失败，返回错误原信息：" + err.Error()
	}

	var errMsg string
	if err != nil {
		if preStr == "" {
			preStr = defaultErrMsg
		}
		errMsg = fmt.Sprintf("%s \n\tat %s:%d\nCause by: %s\n", preStr, f.Name(), line, err.Error())
	}
	return errMsg
}

/**
 * 获得错误描述的同时携带上代码堆栈信息
 */
func GetErrorStackF(err error, preStrFmt string, a ...interface{}) string {
	preStr := ""
	if preStrFmt == "" {
		preStr = defaultErrMsg
	} else {
		preStr = fmt.Sprintf(preStrFmt, a...)
	}
	pc, _, line, ok := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	if !ok {
		return "GetErrorStack 方法获取堆栈失败，返回错误原信息：" + err.Error()
	}

	var errMsg string
	if err != nil {
		if preStr == "" {
			preStr = defaultErrMsg
		}
		errMsg = fmt.Sprintf("%s \n\tat %s:%d\nCause by: %s\n", preStr, f.Name(), line, err.Error())
	}
	return errMsg
}
