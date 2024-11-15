package warp

import "fmt"

type Hybrid struct {
	stream   *Stream
	datagram *Datagram
	server   *Server
	addr     string
}

func NewHybrid(stream *Stream, datagram *Datagram, server *Server) Hybrid {
	h := Hybrid{
		stream:   stream,
		datagram: datagram,
		server:   server,
		addr:     datagram.inner.RemoteAddr().String(),
	}

	return h
}

func (h *Hybrid) Write(buf []byte) (n int, err error) {
	expectedFragments := (len(buf) + 1249) / 1250 // ceil(len(buf) / 1250)
	threshold := int(h.server.GetPacketThreshold(h.addr))
	fmt.Printf("** Fragments: %d, Threshold: %d, In queue: %d\n", expectedFragments, threshold, h.datagram.fragmentToSend)
	if (expectedFragments + h.datagram.fragmentToSend) > threshold {
		fmt.Println("** Using stream")
		return h.stream.Write(buf)
	} else {
		fmt.Println("** Using datagram")
		return h.datagram.Write(buf)
	}
}
