package main

import (
	"log"
	"node/api"
	"node/blockchain"
	"node/config"
	"node/memory"
	"node/storage"
	"node/websocket/reader"
	"os"
	"os/signal"
	"syscall"
)

func waitForSignal() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}

func Init() {
	config.Init()
	memory.Init()
	storage.Init()

	log.Println("System init success")
}

func main() {
	Init()

	go reader.Init()
	go api.ServerStart()
	go blockchain.Init()

	waitForSignal()
}
