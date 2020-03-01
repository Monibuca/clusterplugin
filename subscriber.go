package clusterplugin

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Monibuca/engine/avformat"

	. "github.com/Monibuca/engine"
	"github.com/Monibuca/engine/pool"
)

func ListenBare(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if MayBeError(err) {
		return err
	}
	var tempDelay time.Duration

	for {
		conn, err := listener.Accept()
		println(conn.RemoteAddr().String())
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				println("bare: Accept error: " + err.Error() + "; retrying in " + strconv.FormatFloat(tempDelay.Seconds(), 'f', 2, 64))
				// fmt.Printf("bare: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}

		tempDelay = 0

		go process(conn)
	}
}

func process(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	connAddr := conn.RemoteAddr().String()
	stream := OutputStream{
		SendHandler: func(p *avformat.SendPacket) error {
			head := pool.GetSlice(9)
			head[0] = p.Packet.Type - 7
			binary.BigEndian.PutUint32(head[1:5], p.Timestamp)
			binary.BigEndian.PutUint32(head[5:9], uint32(len(p.Packet.Payload)))
			if _, err := conn.Write(head); err != nil {
				return err
			}
			pool.RecycleSlice(head)
			if _, err := conn.Write(p.Packet.Payload); err != nil {
				return err
			}
			return nil
		}, SubscriberInfo: SubscriberInfo{
			ID:   connAddr,
			Type: "Bare",
		},
	}
	for {
		cmd, err := reader.ReadByte()
		if err != nil {
			return
		}
		bytes, err := reader.ReadBytes(0)
		if err != nil {
			return
		}
		bytes = bytes[0 : len(bytes)-1]
		switch cmd {
		case MSG_SUBSCRIBE:
			if stream.Room != nil {
				fmt.Printf("bare stream already exist from %s", conn.RemoteAddr())
				return
			}
			go stream.Play(string(bytes))
		case MSG_AUTH:
			sign := strings.Split(string(bytes), ",")
			head := []byte{MSG_AUTH, 2}
			if len(sign) > 1 && AuthHooks.Trigger(sign[1]) == nil {
				head[1] = 1
			}
			conn.Write(head)
			conn.Write(bytes[0 : len(bytes)+1])
		case MSG_SUMMARY: //收到从服务器发来报告，加入摘要中
			summary := &ServerSummary{}
			if err = json.Unmarshal(bytes, summary); err == nil {
				summary.Address = connAddr
				Summary.Report(summary)
				if _, ok := slaves.Load(connAddr); !ok {
					slaves.Store(connAddr, conn)
					if Summary.Running() {
						orderReport(io.Writer(conn), true)
					}
					defer slaves.Delete(connAddr)
				}
			}
		default:
			fmt.Printf("bare receive unknown cmd:%d from %s", cmd, conn.RemoteAddr())
			return
		}
	}
}
