package brutehash

import (
	"../redisworker"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"math/rand"
	"time"
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
	for {
		select {
		case <-hw.tick.C:
			hw.privKey = generatePrivKey()
		default:
			incrementPrivKey(hw.privKey)
			priv := convertToPrivateKey(hw.privKey)
			address := crypto.PubkeyToAddress(priv.PublicKey).Hex()
			hw.redisBroker.SavePrivateKey(redisworker.EthKey{Address:address, PrivateKey:fmt.Sprintf("%x", priv.D)})
		}
	}
}

func (hw *HashWorker)runRand()  {
	for {
		select {
		case <-hw.tick.C:
			hw.privKey = generatePrivKey()
		default:
			incrementKeyRand(hw.privKey)
			priv := convertToPrivateKey(hw.privKey)
			address := crypto.PubkeyToAddress(priv.PublicKey).Hex()
			hw.redisBroker.SavePrivateKey(redisworker.EthKey{Address:address, PrivateKey:fmt.Sprintf("%x", priv.D)})
		}
	}
}
