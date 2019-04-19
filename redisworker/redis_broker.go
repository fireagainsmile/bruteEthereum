package redisworker

import (
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"os"
	"sync"
)
var (
	 Worker = 4
)

type EthKey struct {
	Address string
	PrivateKey string
}

type RedisBroker struct {
	c *redis.Client
	stopChannel chan bool
	MessageChannel chan EthKey
	HashChannel chan  EthKey
	workWG sync.WaitGroup
}
var logChannel = make(chan EthKey, 2)

func NewRedisBroker(uri string) *RedisBroker {
	rb := &RedisBroker{
		c:redis.NewClient(&redis.Options{Addr:uri}),
		MessageChannel:make(chan EthKey, Worker * 2),
		HashChannel:make(chan  EthKey, Worker * 3),
	}
	go rb.loop()
	go monitorLog()
	return rb
}

func (rb *RedisBroker)store(key EthKey) error {
	//val, err := conn.Do("GET", key.Address)
	val,err := rb.c.Do("EXISTS", key.Address).Bool()
	if err != nil {
		log.Print("[ERROR]: failed to get record from redis", err)
		return  err
	}
	if val != false{
		return nil
	}
	_, err = rb.c.Do("SET", key.Address, "").String()
	return err
}

func (rb *RedisBroker)loop(){
	for {
		select {
		case <-rb.stopChannel:
			return
		case message := <- rb.MessageChannel:
			rb.store(message)
		case hashed := <-rb.HashChannel:
			rb.savePrivateKey(hashed)
		}
	}
}

func (rb *RedisBroker)SaveAccount(ek EthKey)  {
	rb.MessageChannel <- ek
}

func (rb *RedisBroker)savePrivateKey(ek EthKey) error {
	exist, err := rb.c.Do("exists", ek.Address).Bool()
	if err != nil {
		log.Println("[ERROR] Encountered an error when querying redis",err)
		return err
	}
	if exist != true{
		return nil
	}
	_, err = rb.c.Do("set", ek.Address, ek.PrivateKey).Bool()
	log.Println("[SUCCESS] save private key",ek.PrivateKey)
	logChannel <- ek
	return err
}
func (rb *RedisBroker)SavePrivateKey(ek EthKey) error {
	val, err := rb.c.Do("EXISTS", ek.Address).Bool()
	if err != nil {
		return err
	}
	if val {
		_, err = rb.c.Do("set", ek.Address, ek.PrivateKey).Bool()
		logChannel <- ek
		return err
	}
	return nil
}

func writeToFile(ek EthKey) error {
	input, err := os.OpenFile("priv.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer input.Close()
	_, err = input.WriteString(fmt.Sprintf("keys: %s, %s \n", ek.Address, ek.PrivateKey) )
	return err
}
func monitorLog()  {
	for  {
		select {
		case ek := <- logChannel:
			writeToFile(ek)
		}
	}
}
