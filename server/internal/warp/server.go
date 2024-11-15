package warp

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/TA-MoQ/quic-go"
	"github.com/TA-MoQ/quic-go/http3"
	"github.com/TA-MoQ/quic-go/logging"
	"github.com/TA-MoQ/webtransport-go"
	"github.com/kixelated/invoker"
)

type Server struct {
	inner *webtransport.Server
	media *Media

	// The following properties were added to implement tc rate limiting
	// tcRate is the Mbps value which is read from a file.
	// continueStreaming is a boolean which is set when a user pauses or plays the video
	// these two variables are server-scoped meaning they affect other sessions as well.
	// Hence, the tests should be conducted by one user.
	tcRate            float64
	isTcActive        bool
	continueStreaming bool

	// Whether the DASH file is for streaming or not
	isStreaming bool

	sessions    invoker.Tasks
	connIdAddr  map[quic.ConnectionID]string
	sessionCwnd map[string]int64

	rwMutex       sync.RWMutex
}

type ServerConfig struct {
	Addr        string
	Cert        *tls.Certificate
	LogDir      string
	IsStreaming bool
}

func NewServer(config ServerConfig, media *Media) (s *Server, err error) {
	s = new(Server)

	s.connIdAddr = map[quic.ConnectionID]string{}
	s.sessionCwnd = map[string]int64{}
	s.isStreaming = config.IsStreaming
	s.continueStreaming = true
	s.tcRate = -1

	quicConfig := &quic.Config{
		// Tracer: qlog.DefaultTracer,
	}

	// if config.LogDir != "" {
	// 	quicConfig.Tracer = qlog.NewTracer(func(p logging.Perspective, connectionID []byte) io.WriteCloser {
	// 		path := fmt.Sprintf("%s-%s.qlog", p, hex.EncodeToString(connectionID))

	// 		f, err := os.Create(filepath.Join(config.LogDir, path))
	// 		if err != nil {
	// 			// lame
	// 			panic(err)
	// 		}

	// 		return f
	// 	})
	// }
	// packetsSent := 0
	// packetsLost := 0
	quicConfig.Tracer = func(ctx context.Context, p logging.Perspective, connID quic.ConnectionID) *logging.ConnectionTracer {
		return &logging.ConnectionTracer{
			StartedConnection: func(local, remote net.Addr, src, dest quic.ConnectionID) {
				s.connIdAddr[connID] = remote.String()
				s.SetPacketThreshold(remote.String(), 0)
			},
			Close: func() {
				delete(s.sessionCwnd, s.connIdAddr[connID])
				delete(s.connIdAddr, connID)
			},
			// LostPacket: func(encLevel logging.EncryptionLevel, pn logging.PacketNumber, reason logging.PacketLossReason) {
			// 	packetsLost++
			// 	packetLossRate := (float64(packetsLost) / float64(packetsSent)) * 100
			// 	fmt.Printf("Packet #%d lost with reason %d | Loss Rate: %.2f%%\n", pn, reason, packetLossRate)
			// },
			UpdatedMetrics: func(rttStats *logging.RTTStats, cwnd, bytesInFlight logging.ByteCount, packetsInFlight int) {
				s.SetPacketThreshold(s.connIdAddr[connID], int64(cwnd)/1250)
				// fmt.Println("====================================")
				// fmt.Println("rttStats: ", rttStats)
				// fmt.Println("cwnd: ", cwnd)
				// fmt.Println("bytesInFlight: ", bytesInFlight)
				// fmt.Println("packetsInFlight: ", packetsInFlight)
				// fmt.Println("====================================")
			},
			// BufferedPacket: func(p logging.PacketType, b logging.ByteCount) {
			// 	fmt.Println(p)
			// },
			// SentShortHeaderPacket: func(header *logging.ShortHeader, count logging.ByteCount, ecn logging.ECN, frame *logging.AckFrame, frames []logging.Frame) {
			// 	packetsSent++
			// 	for _, frame := range frames {
			// 		fmt.Printf("[SentShortHeader] Packet #%d ", header.PacketNumber)
			// 		switch f := frame.(type) {
			// 		case *logging.CryptoFrame:
			// 			fmt.Printf("is a CryptoFrame\n")
			// 		case *logging.StreamFrame:
			// 			// fmt.Print(" ", f.StreamID)
			// 			// fmt.Println("This is a StreamFrame", f)
			// 			fmt.Printf("is a StreamFrame with ID: %d\n", f.StreamID)
			// 		case *logging.DatagramFrame:
			// 			fmt.Printf("is a DatagramFrame\n")
			// 		default:
			// 			fmt.Println("is an unknown frame type")
			// 		}
			// 		fmt.Println("Frame Content: ", frame)
			// 	}
			// 	fmt.Print("\n")
			// },
			// SentLongHeaderPacket: func(header *logging.ExtendedHeader, count logging.ByteCount, ecn logging.ECN, frame *logging.AckFrame, frames []logging.Frame) {
			// 	packetsSent++
			// 	// fmt.Println(time.Now())
			// 	// for _, frame := range frames {
			// 	// 	fmt.Printf("[SentLongHeader] Packet #%d ", header.PacketNumber)
			// 	// 	switch f := frame.(type) {
			// 	// 	case *logging.CryptoFrame:
			// 	// 		fmt.Printf("is a CryptoFrame\n")
			// 	// 	case *logging.StreamFrame:
			// 	// 		// fmt.Print(" ", f.StreamID)
			// 	// 		// fmt.Println("This is a StreamFrame", f)
			// 	// 		fmt.Printf("is a StreamFrame with ID: %d\n", f.StreamID)
			// 	// 	case *logging.DatagramFrame:
			// 	// 		fmt.Printf("is a DatagramFrame\n")
			// 	// 	default:
			// 	// 		fmt.Println("is an unknown frame type")
			// 	// 	}
			// 	// }
			// },
		}
	}
	// quicConfig.Tracer = qlog.DefaultTracer

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*config.Cert},
		//NextProtos:   []string{"h3", "h3-32", "h3-31", "h3-30", "h3-29"},
	}

	mux := http.NewServeMux()

	s.inner = &webtransport.Server{
		H3: http3.Server{
			TLSConfig:  tlsConfig,
			QuicConfig: quicConfig,
			Addr:       config.Addr,
			Handler:    mux,
		},
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	s.media = media

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hijacker, ok := w.(http3.Hijacker)
		if !ok {
			panic("unable to hijack connection: must use kixelated/quic-go")
		}

		conn := hijacker.Connection()

		s.isTcActive = true
		s.tcRate = -1 // reset tc

		// wait for 1 sec for the tc limiting to be applied
		time.Sleep(time.Second)

		fmt.Printf("isTcActive: %t rate: %f\n", s.isTcActive, s.tcRate)

		sess, err := s.inner.Upgrade(w, r)
		if err != nil {
			http.Error(w, "failed to upgrade session", 500)
			return
		}

		err = s.serve(r.Context(), conn, sess)
		if err != nil {
			log.Println(err)
		}
	})

	mux.HandleFunc("/testgcp", func(w http.ResponseWriter, r *http.Request) {
		//sending ping to the client
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"message": "pong"}`))
		if err != nil {
			return
		}
	})

	//mux.HandleFunc("/test-datagram", func(w http.ResponseWriter, r *http.Request) {
	//	sess, err := s.inner.Upgrade(w, r)
	//	if err != nil {
	//		http.Error(w, "failed to upgrade session", 500)
	//		return
	//	}
	//
	//	count := 1
	//	for {
	//		msg, err := sess.ReceiveDatagram(r.Context())
	//		if err != nil {
	//			fmt.Println("SessionDatagram closed, ending datagram listener:", err)
	//			break
	//		}
	//		fmt.Printf("Received datagram %d: %s\n", count, msg)
	//
	//		sendMsg := bytes.ToUpper(msg)
	//		fmt.Printf("Sending datagram %d: %s\n", count, sendMsg)
	//		err = sess.SendDatagram(sendMsg)
	//		if err != nil {
	//			fmt.Println("Error when Sending Datagram, Error: ", err)
	//			return
	//		}
	//		count++
	//	}
	//})
	//
	//mux.HandleFunc("/test-datagram-2", func(w http.ResponseWriter, r *http.Request) {
	//	sess, err := s.inner.Upgrade(w, r)
	//	if err != nil {
	//		http.Error(w, "failed to upgrade session", 500)
	//		return
	//	}
	//
	//	count := 1
	//	msg, err := sess.ReceiveDatagram(r.Context())
	//	if err != nil {
	//		fmt.Println("SessionDatagram closed, ending datagram listener:", err)
	//	}
	//	fmt.Printf("Received datagram %d: %s\n", count, msg)
	//	for {
	//		fmt.Printf("Sending datagram %d: %s\n", count, msg)
	//		err = sess.SendDatagram(msg)
	//		if err != nil {
	//			fmt.Println("Error when Sending Datagram, Error: ", err)
	//			return
	//		}
	//		count++
	//	}
	//})
	//
	//mux.HandleFunc("/test-stream", func(w http.ResponseWriter, r *http.Request) {
	//	sess, err := s.inner.Upgrade(w, r)
	//	if err != nil {
	//		http.Error(w, "failed to upgrade session", 500)
	//		return
	//	}
	//	stream, err := sess.AcceptStream(r.Context())
	//	if err != nil {
	//		fmt.Println("SessionDatagram closed, ending stream listener:", err)
	//	}
	//	count := 1
	//	for {
	//		buf := make([]byte, 3)
	//		n, err2 := stream.Read(buf)
	//		if err2 != nil {
	//			fmt.Println("Nothing to read, Error: ", err)
	//			return
	//		}
	//		msg := buf[:n]
	//		fmt.Printf("Received stream %d: %s\n", count, msg)
	//
	//		sendMsg := bytes.ToUpper(msg)
	//		fmt.Printf("Sending stream %d: %s\n", count, sendMsg)
	//		_, err3 := stream.Write(sendMsg)
	//		if err3 != nil {
	//			fmt.Println("Error when Sending Stream, Error: ", err)
	//			return
	//		}
	//		count++
	//	}
	//})

	return s, nil
}

func (s *Server) GetPacketThreshold(addr string) int64 {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	return s.sessionCwnd[addr]
}

func (s *Server) SetPacketThreshold(addr string, threshold int64) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	s.sessionCwnd[addr] = threshold
}

func (s *Server) runTcProfile(ctx context.Context) (err error) {
	// profiles: profile_cascade, profile_lte, profile_twitch
	// set profile name to the one of the options above to run tc netem.
	// TODO: get initial value from a configuration file, allow player to set this remotely
	profile_name := ""
	if profile_name == "" {
		return nil
	}

	data, err := ioutil.ReadFile("./tc_scripts/" + profile_name)
	if err != nil {
		fmt.Println(err.Error())
		log.Fatal(err)
	}

	lines := strings.Split(string(data), "\n")
	i := -1
	for i = 0; i < len(lines); i++ {
		// don't change tc rate if streaming is paused
		if !s.continueStreaming {
			i = i - 1
			time.Sleep(time.Millisecond * 50)
			continue
		}

		// -1 means, reset tc
		if s.tcRate == -1.0 {
			fmt.Printf("resetting tc | isTcActive: %t rate: %f\n", s.isTcActive, s.tcRate)
			cmd := exec.Command("bash", "./tc_scripts/tc_reset.sh")
			stdout, err := cmd.Output()
			if err == nil {
				fmt.Printf("tc reset output: %s", string(stdout))
			} else {
				fmt.Println(err.Error())
				return err
			}
			s.tcRate = 0
		}

		if !s.isTcActive {
			// reset line counter
			i = -1
			time.Sleep(time.Millisecond * 100)
			continue
		}

		line := lines[i]
		parts := strings.Split(line, " ")

		if len(parts) == 2 {

			action := parts[0]
			value := parts[1]

			if action == "rate" {
				fmt.Printf("rate %s\n", value)

				if err == nil {
					float_val, err := strconv.ParseFloat(value, 64)

					if err == nil {
						s.tcRate = float_val / 1024 // Mbps

						// pass Mpbs to the script
						cmd := exec.Command("bash", "./tc_scripts/throttle.sh", fmt.Sprintf("%.1f", s.tcRate))
						stdout, err := cmd.Output()
						if err == nil {
							fmt.Printf("tc command: %s", string(stdout))
						} else {
							fmt.Println(err.Error())
							return err
						}

					} else {
						continue
					}
				} else {
					fmt.Println(err.Error())
					return err
				}

			} else if action == "wait" {
				fmt.Printf("wait %s\n", value)
				if err == nil {
					float_val, err := strconv.ParseFloat(value, 64)
					if err == nil {
						passed_duration_ms := 0.0
						sleep_interval := 10
						for passed_duration_ms < float_val*1000 {
							// if stream is paused, hold tc rate
							if s.continueStreaming {
								passed_duration_ms += float64(sleep_interval)
							}
							err = invoker.Sleep(time.Millisecond * time.Duration(sleep_interval))(ctx)
							if err != nil {
								fmt.Println(err.Error())
								return err
							}

						}

					} else {
						continue
					}
				} else {
					fmt.Println(err.Error())
					return err
				}
			} else {
				continue
			}
		}

		if i == len(lines)-1 {
			// loop
			i = -1
		}
	}
	return nil
}

func (s *Server) runServe(ctx context.Context) (err error) {
	return s.inner.ListenAndServe()
}

func (s *Server) runShutdown(ctx context.Context) (err error) {
	<-ctx.Done()
	s.inner.Close()
	return ctx.Err()
}

func (s *Server) Run(ctx context.Context) (err error) {
	return invoker.Run(ctx, s.runServe, s.runTcProfile, s.runShutdown, s.sessions.Repeat)
}

func (s *Server) serve(ctx context.Context, conn quic.Connection, sess *webtransport.Session) (err error) {
	defer func() {
		if err != nil {
			sess.CloseWithError(1, err.Error())
		} else {
			sess.CloseWithError(0, "end of broadcast")
		}
	}()

	ss, err := NewSession(conn, sess, s.media, s)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	err = ss.Run(ctx)
	if err != nil {
		return fmt.Errorf("terminated session: %w", err)
	}

	return nil
}
