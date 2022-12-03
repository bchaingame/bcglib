package bcg

import (
	"net"
)

func TcpConnect(ip string) (net.Conn, error) {
	return net.Dial("tcp", ip)
}

type TcpOnConnect func(conn net.Conn)

func TcpStartServer(port string, cb TcpOnConnect) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		LogRed("Error listening", err.Error())
		return
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				LogRed("Error accepting", err.Error())
				break
			}
			go cb(conn)
		}
	}()
}
