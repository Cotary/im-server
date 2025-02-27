package epoll_server

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"im-server/util"
	"io"
	"math"
	"strings"
)

const (
	//连续帧
	ContinuousMessage = 0
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

type websocketClient struct {
	SecWebsocketKey string
}

// 解析数据
func decodeData(e *event, reqContext []byte) []byte {
	var content []byte
	if len(reqContext) == 0 {
		return content
	}

	if e.websocket == nil {
		//判断是不是websocket ,有没有连接，没有就升级并且连接
		if isWebsocketConnect(e, reqContext) {
			upgrade(e, reqContext)
		}
	} else {
		var frameType byte
		var err error
		frameType, content, err = ReadIframe(reqContext)
		if err != nil {
			util.LogPrintln("ReadIframe   读数据失败")
			return content
		}
		if frameType == TextMessage {
			return content
			//WriteIframe(s, content, TextMessage)
		} else if frameType == PingMessage {
			pongData := GetIframeByte(e, []byte{}, PongMessage)
			e.writeSocket(pongData)

		}
	}
	return content
}

func isWebsocketConnect(s *event, content []byte) bool {
	isHttp := false
	// 先暂时这么判断
	if string(content[0:3]) == "GET" {
		isHttp = true
	}

	if isHttp {
		headers := parseHandshake(string(content))
		secWebsocketKey := headers["Sec-WebSocket-Key"]
		if len(secWebsocketKey) > 0 {
			s.websocket = &websocketClient{
				SecWebsocketKey: secWebsocketKey,
			}
			return true
		}

	}
	return false
}

func parseHandshake(content string) map[string]string {
	headers := make(map[string]string, 10)
	lines := strings.Split(content, "\r\n")
	for _, line := range lines {
		if len(line) >= 0 {
			words := strings.Split(line, ":")
			if len(words) == 2 {
				headers[strings.Trim(words[0], " ")] = strings.Trim(words[1], " ")
			}
		}
	}
	return headers
}

func upgrade(s *event, content []byte) {
	secWebsocketKey := s.websocket.SecWebsocketKey
	// NOTE：这里省略其他的验证
	guid := "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	// 计算Sec-WebSocket-Accept
	h := sha1.New()
	io.WriteString(h, secWebsocketKey+guid)
	accept := make([]byte, 28)
	base64.StdEncoding.Encode(accept, h.Sum(nil))
	util.LogPrintln(string(accept))
	response := "HTTP/1.1 101 Switching Protocols\r\n"
	response = response + "Sec-WebSocket-Accept: " + string(accept) + "\r\n"
	response = response + "Connection: Upgrade\r\n"
	response = response + "Upgrade: websocket\r\n\r\n"
	//握手，升级协议
	s.writeSocket([]byte(response))
	util.LogPrintln("升级协议")
}
func pollByte(context *[]byte, num int) (res []byte) {
	res = (*context)[0:num]
	*context = (*context)[num:]
	return
}

func ReadIframe(context []byte) (frameType byte, data []byte, err error) {

	tempContext := make([]byte, len(context))
	copy(tempContext, context)
	err = nil
	//第一个字节：FIN + RSV1-3 + OPCODE
	opcodeByte := pollByte(&tempContext, 1)
	//二进制运算符&通过对两个操作数一位一位的比较产生一个新的值，对于每个位，只有两个操作数的对应位都为1时结果才为1.如10000001&11000000的结果为“10000000”
	//<<左移 *2 >>右移/2  ，左移是可以保留的，比如0001左移4，就是00010000，右移4就是0000，就没有了

	FIN := opcodeByte[0] >> 7 //只保留第一位
	//RSV1 := opcodeByte[0] >> 6 & 1 //保留左边两位，在和1相与
	//RSV2 := opcodeByte[0] >> 5 & 1
	//RSV3 := opcodeByte[0] >> 4 & 1
	frameType = opcodeByte[0] & 0xf //0xf就是15,1111
	//util.LogPrintln(FIN, RSV1, RSV2, RSV3, frameType)

	payloadLenByte := pollByte(&tempContext, 1)

	payloadLen := int(payloadLenByte[0] & 0x7F)
	mask := payloadLenByte[0] >> 7
	//处理长度
	textLen := 0
	if payloadLen <= 125 {
		textLen = payloadLen
	} else if payloadLen == 126 {
		extendedByte := pollByte(&tempContext, 2)
		textLen = int(binary.BigEndian.Uint16(extendedByte))
	} else if payloadLen == 127 {
		extendedByte := pollByte(&tempContext, 8)
		textLen = int(binary.BigEndian.Uint64(extendedByte))
	}
	//处理mask
	var maskingByte []byte
	if mask == 1 {
		maskingByte = pollByte(&tempContext, 4)
	}

	//得到荷载数据

	payloadDataByte := pollByte(&tempContext, textLen)

	//荷载和掩码计算
	dataByte := make([]byte, textLen)
	for i := 0; i < textLen; i++ {
		if mask == 1 {
			dataByte[i] = payloadDataByte[i] ^ maskingByte[i%4]
		} else {
			dataByte[i] = payloadDataByte[i]
		}
	}
	//util.LogPrintln("原始帧：")
	//fmt.Printf("%x ", readBuff)  转换成16进制
	if FIN == 1 {
		data = dataByte
		return
	}
	_, nextData, err := ReadIframe(context)
	if err != nil {
		return
	}
	data = append(data, nextData...)
	return
}

// 得到写websocket 发送格式
func GetIframeByte(e *event, data []byte, opCode byte) (sendBuff []byte) {
	iframePerByte := 1000 //最大帧传多少数据
	var maskingKey []byte

	//控制帧
	if len(data) == 0 && opCode != TextMessage {
		sendBuff = SingleIframe(e, data, true, opCode, maskingKey)
		return
	}

	//数据帧

	//分片
	//1.一个未分片的消息只有一帧（FIN为1，opcode非0）
	//2.一个分片的消息由起始帧（FIN为0，opcode非0），若干（0个或多个）帧（FIN为0，opcode为0），结束帧（FIN为1，opcode为0）。
	if len(data) > iframePerByte {
		countIframe := int(math.Ceil(float64(len(data)) / float64(iframePerByte))) //要划分多少帧(向上取整)  2.5 1 3
		//这个为nil ,maskingKey:= []byte{} 就不等于nil
		for i := 0; i < countIframe; i++ {
			//0 1 2
			start := i * iframePerByte
			end := int(math.Min(float64((i+1)*iframePerByte), float64(len(data))))

			singleData := data[start:end]
			isFin := i == (countIframe - 1)
			opCode = ContinuousMessage
			if i == 0 {
				opCode = TextMessage
			}
			//注意这个切片，start是开始的下标，end是实际结束的下标+1
			//util.LogPrintln("循环：", i, start, end, data[start:end], isFin)
			tempBuff := SingleIframe(e, singleData, isFin, opCode, maskingKey)
			sendBuff = append(sendBuff, tempBuff...)
		}
	} else {
		//没有数据的，应该
		sendBuff = SingleIframe(e, data, true, TextMessage, maskingKey)
	}

	return
}

//0                   1                   2                   3
//0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//+-+-+-+-+-------+-+-------------+-------------------------------+
//|F|R|R|R| opcode|M| Payload len |    Extended payload length    |
//|I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
//|N|V|V|V|       |S|             |   (if payload len==126/127)   |
//| |1|2|3|       |K|             |                               |
//+-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
//|     Extended payload length continued, if payload len == 127  |
//+ - - - - - - - - - - - - - - - +-------------------------------+
//|                               |Masking-key, if MASK set to 1  |
//+-------------------------------+-------------------------------+
//| Masking-key (continued)       |          Payload Data         |
//+-------------------------------- - - - - - - - - - - - - - - - +
//:                     Payload Data continued ...                :
//+ - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
//|                     Payload Data continued ...                |
//+---------------------------------------------------------------+

func SingleIframe(e *event, data []byte, isFin bool, opCode byte, maskingKey []byte) []byte {
	//65536
	//18446744073709552000

	//处理数据，进行掩码计算
	//var maskingKey []byte  //这个为nil ,maskingKey:= []byte{} 就不等于nil
	length := len(data)
	util.LogPrintln("长度是:", length)
	maskedData := make([]byte, length)
	for i := 0; i < length; i++ {
		//先不返回掩码
		if maskingKey != nil {
			maskedData[i] = data[i] ^ maskingKey[i%4]
		} else {
			maskedData[i] = data[i]
		}
	}

	var sendBuff []byte

	//b1

	//fin
	fin := byte(0x0)
	if isFin {
		fin = byte(0x80)
	}
	//opcode
	b1 := fin | opCode
	sendBuff = append(sendBuff, b1)

	//b2

	//看有没有掩码
	b2 := byte(0) //00000000
	if maskingKey != nil {
		b2 |= 1 << 7 //10000000
	}

	//payload len

	//extended payload len
	payloadLen := byte(0)

	if length >= 65536 {
		payloadLen = 127
		sendBuff = append(sendBuff, b2|payloadLen)

		extendedPayloadLen := make([]byte, 8)
		binary.BigEndian.PutUint64(extendedPayloadLen, uint64(length)) //64位转为二进制的8个字节
		sendBuff = append(sendBuff, extendedPayloadLen...)

	} else if length > 125 {
		payloadLen = 126
		sendBuff = append(sendBuff, b2|payloadLen)

		extendedPayloadLen := make([]byte, 2)
		binary.BigEndian.PutUint16(extendedPayloadLen, uint16(length))
		sendBuff = append(sendBuff, extendedPayloadLen...)

	} else {
		payloadLen = byte(length)
		sendBuff = append(sendBuff, b2|payloadLen)
	}

	//Masking-key
	if maskingKey != nil {
		sendBuff = append(sendBuff, maskingKey...)

	}

	//写数据
	sendBuff = append(sendBuff, maskedData...)
	return sendBuff
}
