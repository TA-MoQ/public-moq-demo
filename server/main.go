package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"

	"github.com/kixelated/invoker"
	"github.com/kixelated/warp-demo/server/internal/warp"
)

func main() {
	err := run(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) (err error) {
	addr := flag.String("addr", ":4443", "HTTPS server address")
	//server deploy env
	cert := flag.String("tls-cert", "/etc/letsencrypt/live/dickyarian.blue/fullchain.pem", "TLS certificate file path")
	key := flag.String("tls-key", "/etc/letsencrypt/live/dickyarian.blue/privkey.pem", "TLS certificate file path")

	//local env
	//cert := flag.String("tls-cert", "../cert/localhost.crt", "TLS certificate file path")
	//key := flag.String("tls-key", "../cert/localhost.key", "TLS certificate file path")
	logDir := flag.String("log-dir", "", "logs will be written to the provided directory")

	dash := flag.String("dash", "../media/playlist.mpd", "DASH playlist path")
	//dash := flag.String("dash", "C:/Users/Farrel/Documents/Kuliah/SEM-8/Tugas Akhir/Repositories/test-av1/playlist.mpd", "DASH playlist path")

	isStreaming := flag.Bool("streaming", false, "If the dash file meant for streaming")

	flag.Parse()

	media, err := warp.NewMedia(*dash, *isStreaming)
	if err != nil {
		return fmt.Errorf("failed to open media: %w", err)
	}

	tlsCert, err := tls.LoadX509KeyPair(*cert, *key)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	config := warp.ServerConfig{
		Addr:        *addr,
		Cert:        &tlsCert,
		LogDir:      *logDir,
		IsStreaming: *isStreaming,
	}

	ws, err := warp.NewServer(config, media)
	if err != nil {
		return fmt.Errorf("failed to create warp server: %w", err)
	}

	log.Printf("listening on %s", *addr)

	return invoker.Run(ctx, invoker.Interrupt, ws.Run)
}
