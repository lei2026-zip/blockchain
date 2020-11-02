package blockchain

import (
	"github.com/boltdb/bolt"
	"fmt"
	"errors"
)

//桶的名称，该桶用于装区块信息
var BUCKET_NAME = "blocks"
//表示最新的区块的key名
var LAST_KEY = "lasthash"
//存储区块数据的文件
var CHAINDB = "chain.db"

/**
 * 区块链结构体实例定义:用于表示代表一条区块链
 * 该区块链包含以下功能:
 		① 将新产生的区块与已有的区块链接起来，并保存
		② 可以查询某个区块的信息
		③ 可以将所有区块进行遍历，输出区块信息
 */
type BlockChain struct {
	LastHash []byte //最新区块hash
	BoltDb   *bolt.DB
}

/**
 * 用于创建一条区块链，并返回该区块链实例
	解释：由于区块链就是由一个一个的区块组成的，因此,如果要创建一条区块链，那么必须要先
		创建一个区块，该区块作为该条区块链的创世区块。
 */
func NewBlockChain() BlockChain {
	//0、打开存储区块数据的chain.db文件
	db, err := bolt.Open(CHAINDB, 0600, nil)
	if err != nil {
		panic(err.Error())
	}
	var bl BlockChain
	//先从区块链中都看是否创世区块已经存在
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_NAME))
		if bucket == nil {
			bucket, err = bucket.CreateBucket([]byte(BUCKET_NAME))
			if err != nil {
				panic(err.Error())
			}
		}
		lastHash := bucket.Get([]byte(LAST_KEY))
		if len(lastHash) == 0 { //没有创世区块
			//1、创建创世区块
			genesis := CreateGenesisBlock() //创世区块
			//2、创建一个存储区块数据的文件
			fmt.Printf("genesis的Hash值:%x\n", genesis.Hash)
			bl = BlockChain{
				LastHash: genesis.Hash,
				BoltDb:   db,
			}
			genesisBytes, _ := genesis.Serialize()
			bucket.Put(genesis.Hash, genesisBytes)
			bucket.Put([]byte(LAST_KEY), genesis.Hash)
		} else { //有创世区块
			lastHash := bucket.Get([]byte(LAST_KEY))
			lastBlockBytes := bucket.Get(lastHash) //创世区块的[]byte
			lastBlock, err := DeSerialize(lastBlockBytes)
			if err != nil {
				panic("读取区块链数据失败")
			}
			bl = BlockChain{
				LastHash: lastBlock.Hash,
				BoltDb:   db,
			}
		}
		return nil
	})
	return bl
}

/**
 * 调用BlockChain的该SaveBlock方法，该方法可以将一个生成的新区块保存到chain.db文件中
 */
func (bc BlockChain) SaveData(data []byte) (Block, error) {
	db := bc.BoltDb
	var e error
	var lastBlock *Block
	//先查询chain.db中存储的最新的区块
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_NAME))
		if bucket == nil {
			e = errors.New("boltdb未创建，请重试!")
			return e
		}
		//lastHash := bucket.Get([]byte(LAST_KEY))
		lastBlockBytes := bucket.Get(bc.LastHash)
		lastBlock, _ = DeSerialize(lastBlockBytes)
		return nil
	})

	//1、先生成一个区块 把data存入到新生成的区块中
	newBlock := NewBlock(lastBlock.Height+1, data, lastBlock.Hash)

	//更新chain.db 把newblock存入到boltdb中
	db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BUCKET_NAME))
		//key=hash, value=block的byte
		//区块序列化
		newBlockBytes, _ := newBlock.Serialize()
		//把区块信息保存到boltdb中
		bucket.Put(newBlock.Hash, newBlockBytes)
		//更新代表最后一个区块hash值的记录
		bucket.Put([]byte(LAST_KEY), newBlock.Hash)
		bc.LastHash = newBlock.Hash
		return nil
	})
	return newBlock, e
}
