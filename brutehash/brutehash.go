package brutehash

import (
	"../redisworker"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"math/rand"
	"time"
)

var (
	TotalHashed = new(big.Int)
)

type HashWorker struct {
	redisBroker *redisworker.RedisBroker
	privKey []byte
	tick *time.Ticker
}

func NewHashWorker(redis *redisworker.RedisBroker)  *HashWorker{
	ew := &HashWorker{
		redisBroker:redis,
		privKey:make([]byte,32),
	}
	ew.tick = time.NewTicker(time.Minute *3)
	ew.privKey = generatePrivKey()
	return ew
}

func generatePrivKey() []byte {
	priv := make([]byte, 32)
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	for i := 0; i < 31; i ++ {
		priv[i] = byte(r.Intn(256))
	}
	//log.Printf("%x", priv)
	return priv
}

func convertToPrivateKey(privKey []byte) *ecdsa.PrivateKey {
	return crypto.ToECDSAUnsafe(privKey)
}

func incrementPrivKey(privKey []byte) {
	for i := 31; i >= 0; i-- {
		if privKey[i]+1 == 255 {
			privKey[i] = 0
		} else {
			privKey[i] += 1
			break
		}
	}
}

func incrementKeyRand(privKey []byte)  {
	index := rand.Intn(32)
	if privKey[index] + 1 ==255 {
		privKey[index] = 0
		if index == 0 {
			privKey[index] = byte(rand.Intn(255))
			return
		}else {
			privKey[index -1 ] += 1
			return
		}
	}else {
		privKey[index] += 1
		return
	}
}

func (hw *HashWorker)Run()  {
	go hw.run()
}

func (hw *HashWorker)RunRand()  {
	go hw.runRand()
}

func (hw *HashWorker)run()  {
	defer hw.tick.Stop()
	counter := 0
	for {
		select {
		case <-hw.tick.C:
			TotalHashed.Add(TotalHashed, big.NewInt(int64(counter)))
			hw.privKey = generatePrivKey()
			log.Println("Hash Rate:",counter, "total calculated:", TotalHashed.Uint64())
			counter = 0
		default:
			incrementPrivKey(hw.privKey)
			priv := convertToPrivateKey(hw.privKey)
			address := crypto.PubkeyToAddress(priv.PublicKey).Hex()
			hw.redisBroker.SavePrivateKey(redisworker.EthKey{Address:address, PrivateKey:fmt.Sprintf("%x", priv.D)})
			counter ++
		}
	}
}

func (hw *HashWorker)runRand()  {
	defer hw.tick.Stop()
	counter := 0
	for {
		select {
		case <-hw.tick.C:
			hw.privKey = generatePrivKey()
			TotalHashed.Add(TotalHashed, big.NewInt(int64(counter)))
			log.Println("Rand Hash Rate:",counter, " Found keys:", redisworker.Found)
			counter = 0
		default:
			incrementKeyRand(hw.privKey)
			priv := convertToPrivateKey(hw.privKey)
			address := crypto.PubkeyToAddress(priv.PublicKey).Hex()
			hw.redisBroker.SavePrivateKey(redisworker.EthKey{Address:address, PrivateKey:fmt.Sprintf("%x", priv.D)})
			counter ++
		}
	}
}

func (hw *HashWorker)AllRand()  {
	go hw.allRand()
}

func (hw *HashWorker)allRand()  {
	defer hw.tick.Stop()
	counter := 0
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	for {
		select {
		case <-hw.tick.C:
			TotalHashed.Add(TotalHashed, big.NewInt(int64(counter)))
			log.Println("all Rand Hash Rate:",counter, " Found keys:", redisworker.Found)
			counter = 0
		default:
			for i := 0; i < 31; i ++ {
				hw.privKey[i] = byte(r.Intn(256))
			}
			priv := convertToPrivateKey(hw.privKey)
			address := crypto.PubkeyToAddress(priv.PublicKey).Hex()
			hw.redisBroker.SavePrivateKey(redisworker.EthKey{Address:address, PrivateKey:fmt.Sprintf("%x", priv.D)})
			counter ++
		}
	}
}