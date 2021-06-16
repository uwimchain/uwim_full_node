package websocket

type RequestSign struct {
	SenderIp string
	Sign     []byte
	Address  string
	Data     []byte
}
