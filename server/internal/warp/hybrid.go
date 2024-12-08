package warp

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
	threshold := 25
	if (expectedFragments + h.datagram.fragmentToSend) > threshold {
		return h.stream.Write(buf)
	} else {
		return h.datagram.Write(buf)
	}
}
