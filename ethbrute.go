package main

import (
	"./brutehash"
	"./client"
	"./redisworker"
	"github.com/Unknwon/goconfig"
	"os"
	"sync"
)

func main()  {
	cfg, err := goconfig.LoadConfigFile("config.ini")
	if err != nil{
		os.Exit(-1)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	celeryBroker := redisworker.NewRedisBroker("localhost:6379")
	client := client.NewEthClient(celeryBroker, cfg)
	client.SubscribTxs()
	client.Run()
	hw := brutehash.NewHashWorker(celeryBroker)
	hw.Run()
	hw2 := brutehash.NewHashWorker(celeryBroker)
	hw2.RunRand()
	wg.Wait()
}

