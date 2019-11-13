package main

import (
	"./common"
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	flag.Parse()

	config, err := common.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load config: %s", err)
	}
	if err = config.Check(); err != nil {
		log.Fatalf("Configuration error: %s", err)
	}

	fmt.Println(config.Info())
	r := mux.NewRouter()
	r.HandleFunc("/", handler)

	stop := make(chan os.Signal, 2)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Listening on: ", config.Listen)
	server := http.Server{
		Addr:    config.Listen,
		Handler: r,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Println("Error listen: %s", err)
		}
	}()

	<-stop
	fmt.Println("Finish processing...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Error shundown %s", err)
	}
	defer cancel()
	fmt.Println("Server gracefully stopped.")
}

func handler(_ http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI)
	return
}
