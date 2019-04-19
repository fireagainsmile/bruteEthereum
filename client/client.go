package client

import (
	"../redisworker"
	"context"
	"fmt"
	"github.com/Unknwon/goconfig"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"time"
)
var bigOne = big.NewInt(1)

type EthClient struct {
	c *ethclient.Client
	b *redisworker.RedisBroker
	headerChan chan *types.Header
	sub ethereum.Subscription
	chainId *big.Int
	blockNumber *big.Int
	ticker *time.Ticker
	cfg *goconfig.ConfigFile
}

func NewEthClient(rb *redisworker.RedisBroker, cfg *goconfig.ConfigFile) *EthClient {
	provider, err := cfg.GetValue("ethereum", "provider")
	if err != nil {
		log.Fatal(err)
		return nil
	}

	lastBlockNumber, err := cfg.Int("ethereum", "blockNumber")

	if err != nil{
		log.Println("[ERROR] failed to get block Number", err)
	}

	client, err := ethclient.Dial(provider)
	if err != nil {
		log.Println("[ERROR]: failed to connect eth infura", err)
		return nil
	}
	ec:= &EthClient{
		c: client,
		b: rb,
		headerChan: make(chan *types.Header),
		blockNumber:new(big.Int).SetInt64(int64(lastBlockNumber)),
		cfg:cfg,
	}
	ec.ticker = time.NewTicker( time.Minute * 3)
	ec.chainId,_ = ec.c.NetworkID(context.Background())
	return ec
}

func (ec *EthClient)SubscribTxs() error {
	url, _ := ec.cfg.GetValue("ethereum", "wss")
	client, err := ethclient.Dial(url)
	//client, err := ethclient.Dial("wss://ropsten.infura.io/ws")
	if err != nil {
		log.Println("[ERROR] failed to dia wss server", err)
		return err
	}

	//sub, err := ec.c.SubscribeNewHead(context.Background(), ec.headerChan)
	sub, err := client.SubscribeNewHead(context.Background(), ec.headerChan)
	if err != nil {
		log.Println("[ERROR]: failed to subscribe block head from infura", err)
		return err
	}
	ec.sub = sub
	return nil
}

func (ec *EthClient)Run() error {
	currentHeader := ec.getHeaderByNumber(nil)
	go ec.loop()

	//update local db to the latest state
	diff := new(big.Int)
	diff.Sub(currentHeader.Number, ec.blockNumber)
	if currentHeader == nil{
		log.Println("[ERROR] initial header failed")
		return fmt.Errorf("error with current block")
	}
	go func() {
		for i := 0; i < int(diff.Uint64()); i++   {
			ec.updateAccount(currentHeader)
			currentHeader = ec.getHeaderByNumber(currentHeader.Number.Sub(currentHeader.Number, bigOne))
		}
	}()
	return nil
}

func (ec *EthClient)loop() {
	defer ec.c.Close()
	//counter := 0
	for {
		select {
		case err := <-ec.sub.Err():
			log.Println("[ERROR] loop received error from sub", err)
		case <-ec.ticker.C:
			ec.saveStatus()
		case <-ec.headerChan:
			header, _ := ec.c.HeaderByNumber(context.Background(), nil)
			ec.updateAccount(header)
		}
	}
}

func (ec *EthClient)getHeaderByNumber(n *big.Int) *types.Header {
	header, err := ec.c.HeaderByNumber(context.Background(), n)
	if err != nil {
		log.Println("[ERROR failed to get header block")
		return nil
	}
	return header
}

func (ec *EthClient)updateAccount(head *types.Header)  {
	if head == nil {
		return
	}
	block, err := ec.c.BlockByHash(context.Background(), head.Hash())
	if err != nil {
		log.Println("[ERROR]: failed to get block by hash ", err)
		return
	}
	ec.blockNumber = block.Number()
	for _, tx := range block.Transactions() {
		if tx.To() != nil {
			msg, _ := tx.AsMessage(types.NewEIP155Signer(ec.chainId))
			if msg.From().Hex() != msg.To().Hex() {
				//log.Println("saved tx:",tx.Hash().Hex())
				//log.Println(msg.From().Hex(), msg.To().Hex())
				ec.b.SaveAccount(redisworker.EthKey{Address:msg.From().Hex()})
				ec.b.SaveAccount(redisworker.EthKey{Address:msg.To().Hex()})
			}else {
				ec.b.SaveAccount(redisworker.EthKey{Address:msg.To().Hex()})
			}
		}
	}
}

func (ec *EthClient)saveStatus()  {
	ec.cfg.SetValue("ethereum", "blockNumber", ec.blockNumber.String())
	goconfig.SaveConfigFile(ec.cfg, "config.ini")
}


