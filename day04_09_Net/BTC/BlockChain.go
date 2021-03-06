package BTC

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"
)

type BlockChain struct {
	//Blocks []*Block
	Tip []byte
	DB  *bolt.DB
}

////版本1
//func CreateBlockChainWithGenesisBlock(data string) *BlockChain {
//	//1、先判断数据库是否存在，如果有，从数据库读取
//	if dbExists() {
//		fmt.Println("数据库已经存在。。")
//		//A:打开数据库
//		db, err := bolt.Open(DBNAME, 0600, nil)
//		if err != nil {
//			log.Fatal(err)
//		}
//		//defer db.Close()
//		var blockchain *BlockChain
//		//B:读取数据库
//		err = db.View(func(tx *bolt.Tx) error {
//			//C:打开表
//			b := tx.Bucket([]byte(BLOCKTABLENAME))
//			if b != nil {
//				//D:读取最后一个hash
//				hash := b.Get([]byte("1"))
//				//E:创建blockchain
//				blockchain = &BlockChain{hash, db}
//			}
//			return nil
//		})
//		if err != nil {
//			log.Fatal(err)
//		}
//		return blockchain
//	}
//
//	//2.数据库不存在，说明第一次创建，然后存入到数据库中
//	fmt.Println("数据库不存在。。")
//	//A：创建创世区块
//	genesisBlock := CreateGenesisBlock(data)
//	//B：打开数据库
//	db, err := bolt.Open(DBNAME, 0600, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	//defer db.Close()
//	//C:存入数据表
//	err = db.Update(func(tx *bolt.Tx) error {
//		b, err := tx.CreateBucket([]byte(BLOCKTABLENAME))
//		if err != nil {
//			log.Panic(err)
//		}
//		if b != nil {
//			err = b.Put(genesisBlock.Hash, genesisBlock.Serilalize())
//			if err != nil {
//				log.Panic("创世区块存储有误。。。")
//			}
//			//存储最新区块的hash
//			b.Put([]byte("1"), genesisBlock.Hash)
//		}
//		return nil
//	})
//	if err != nil {
//		log.Panic(err)
//	}
//
//	//返回区块链对象
//	return &BlockChain{genesisBlock.Hash, db}
//}

//版本2
func CreateBlockChainWithGenesisBlock(address string, nodeID string) {
	/*
		格式化数据库的名字
			1、修改数据库的名字："blockchain_%s.db"
			2、根据节点生成数据库的名字
	*/
	DBNAME := fmt.Sprintf(DBNAME, nodeID)

	if dbExists(DBNAME) {
		fmt.Println("数据库已经存在。。。")
		return
	}

	fmt.Println("创建创世区块：", address)
	fmt.Println("数据库不存在。。")

	//创建创世区块
	//先创建coinbase交易
	txCoinBase := NewCoinBaseTransaction(address)
	genesisBlock := CreateGenesisBlock([]*Transaction{txCoinBase})
	//打开数据库
	db, err := bolt.Open(DBNAME, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	//存入数据表
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(BLOCKTABLENAME))
		if err != nil {
			log.Panic(err)
		}
		if b != nil {
			err = b.Put(genesisBlock.Hash, genesisBlock.Serilalize())
			if err != nil {
				log.Panic("创世区块存储有误。。。")
			}
			//存储最新区块的hash
			b.Put([]byte("1"), genesisBlock.Hash)
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (bc *BlockChain) AddBlockToBlockChain(txs []*Transaction) {
	//newBlock := NewBlock(data, prevHash, height)
	//bc.Blocks = append(bc.Blocks, newBlock)
	//1、更新数据库
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		//2、打开表
		b := tx.Bucket([]byte(BLOCKTABLENAME))
		if b != nil {
			//2、根据最新区块的hash读取数据，并反序列化最后一个区块
			blockBytes := b.Get(bc.Tip)
			lastBlock := DeserializeBlock(blockBytes)
			//3、创建新的区块
			newBlock := NewBlock(txs, lastBlock.Hash, lastBlock.Height+1)
			//4、将新的区块序列化并存储
			err := b.Put(newBlock.Hash, newBlock.Serilalize())
			if err != nil {
				log.Panic(err)
			}
			//5、更新最后一个hash值，以及blockchain的tip
			b.Put([]byte("1"), newBlock.Hash)
			bc.Tip = newBlock.Hash
		}
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

//判断数据库是否存在
func dbExists(DBName string) bool {
	if _, err := os.Stat(DBName); os.IsNotExist(err) {
		return false
	}
	return true
}

////2.获取一个迭代器的方法
//func (bc *BlockChain) Iterator() *BlockChainIterator {
//	return &BlockChainIterator{bc.Tip, bc.DB}
//}
//
//func (bc *BlockChain) PrintChains() {
//	//1.获取迭代器对象
//	bcIterator := bc.Iterator()
//
//	var count = 0
//	//2.循环迭代
//	for {
//		block := bcIterator.Next()
//		count++
//		fmt.Printf("第%d个区块的信息：\n", count)
//		//获取当前hash对应的数据，并进行反序列化
//		fmt.Printf("\t高度：%d\n", block.Height)
//		fmt.Printf("\t上一个区块的hash：%x\n", block.PrevBlockHash)
//		fmt.Printf("\t当前的hash：%x\n", block.Hash)
//		fmt.Printf("\t数据：%s\n", block.Data)
//		fmt.Printf("\t时间：%s\n", time.Unix(block.TimeStamp, 0).Format("2006-01-02 15:04:05"))
//		fmt.Printf("\t次数：%d\n", block.Nonce)
//
//		//3.直到父hash值为0
//		hashInt := new(big.Int)
//		hashInt.SetBytes(block.PrevBlockHash)
//		if big.NewInt(0).Cmp(hashInt) == 0 {
//			break
//		}
//	}
//}

func (bc *BlockChain) Iterator() *BlockChainIterator {
	return &BlockChainIterator{bc.Tip, bc.DB}
}

//func (bc *BlockChain) PrintChains() {
//	//1.获取迭代器对象
//	bcIterator := bc.Iterator()
//
//	var count = 0
//	//2.循环迭代
//	for {
//		block := bcIterator.Next()
//		count++
//		fmt.Printf("第%d个区块的信息：\n", count)
//		//获取当前hash对应的数据，并进行反序列化
//		fmt.Printf("\t高度：%d\n", block.Height)
//		fmt.Printf("\t上一个区块的hash：%x\n", block.PrevBlockHash)
//		fmt.Printf("\t当前的hash：%x\n", block.Hash)
//		fmt.Printf("\t数据：%s\n", block.Data)
//		fmt.Printf("\t时间：%s\n", time.Unix(block.TimeStamp, 0).Format("2006-01-02 15:04:05"))
//		fmt.Printf("\t次数：%d\n", block.Nonce)
//
//		//3.直到父hash值为0
//		hashInt := new(big.Int)
//		hashInt.SetBytes(block.PrevBlockHash)
//		if big.NewInt(0).Cmp(hashInt) == 0 {
//			break
//		}
//	}
//}

//打印区块信息
func (bc *BlockChain) PrintChains() {
	//1.获取迭代器对象
	bcIterator := bc.Iterator()

	for {
		block := bcIterator.Next()

		fmt.Printf("第%d个区块的信息：\n", block.Height+1)
		//获取当前hash对应的数据，并进行反序列化
		fmt.Printf("\t高度：%d\n", block.Height)
		fmt.Printf("\t上一个区块的hash：%x\n", block.PrevBlockHash)
		fmt.Printf("\t当前的hash：%x\n", block.Hash)
		//fmt.Printf("\t数据：%s\n", block.Txs)
		fmt.Println("\t交易：")
		for _, tx := range block.Txs {
			fmt.Printf("\t\t交易ID：%x\n", tx.TxID)
			fmt.Println("\t\tVins:")
			for _, in := range tx.Vins {
				fmt.Printf("\t\t\tTxID:%x\n", in.TxID)
				fmt.Printf("\t\t\tVout:%d\n", in.Vout)
				fmt.Printf("\t\t\tScriptSiq:%s\n", in.PublicKey)
			}
			fmt.Println("\t\tVouts:")
			for _, out := range tx.Vouts {
				fmt.Printf("\t\t\tvalue:%dd\n", out.Value)
				fmt.Printf("\t\t\tScriptPubKey:%s\n", out.PubKeyHash)
			}
		}
		fmt.Printf("\t时间：%s\n", time.Unix(block.TimeStamp, 0).Format("2006-01-02 15:04:05"))
		fmt.Printf("\t次数：%d\n", block.Nonce)

		hashInt := new(big.Int)
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(hashInt) == 0 {
			break
		}
	}
}

//获取区块链
func GetBlockchainObject(nodeID string) *BlockChain {
	DBNAME := fmt.Sprintf(DBNAME, nodeID)

	if !dbExists(DBNAME) {
		fmt.Println("数据库不存在，无法获取区块链。。")
		return nil
	}

	db, err := bolt.Open(DBNAME, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var blockchain *BlockChain

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BLOCKTABLENAME))
		if b != nil {
			hash := b.Get([]byte("1"))
			blockchain = &BlockChain{hash, db}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	return blockchain
}

//挖掘新的区块
func (bc *BlockChain) MineNewBlock(from, to, amount []string, nodeID string) {
	var txs []*Transaction

	//奖励
	tx := NewCoinBaseTransaction(from[0])
	txs = append(txs, tx)

	utxoSet := &UTXOSet{bc}

	for i := 0; i < len(from); i++ {
		amountInt, _ := strconv.ParseInt(amount[i], 10, 64)
		tx := NewSimpleTransaction(from[i], to[i], amountInt, utxoSet, txs, nodeID)

		txs = append(txs, tx)
	}

	var block *Block
	var newBlock *Block
	bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BLOCKTABLENAME))
		if b != nil {
			hash := b.Get([]byte("1"))
			blockBytes := b.Get(hash)
			block = DeserializeBlock(blockBytes)
		}
		return nil
	})

	//在建立新区块链，对txs进行签名验证
	_txs := []*Transaction{}
	for _, tx := range txs {
		if bc.VerifyTransaction(tx, _txs) != true {
			log.Panic("签名验证失败。。。")
		}
		_txs = append(_txs, tx)
	}

	newBlock = NewBlock(txs, block.Hash, block.Height+1)

	bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BLOCKTABLENAME))
		if b != nil {
			b.Put(newBlock.Hash, newBlock.Serilalize())
			b.Put([]byte("1"), newBlock.Hash)
			bc.Tip = newBlock.Hash
		}
		return nil
	})
}

//找到所有未花费的交易输出
func (bc *BlockChain) UnUTXOs(address string, txs []*Transaction) []*UTXO {
	var unUTXOs []*UTXO
	spentTxOutputs := make(map[string][]int)

	for i := len(txs) - 1; i >= 0; i-- {
		unUTXOs = caculate(txs[i], address, spentTxOutputs, unUTXOs)
	}

	bcIterator := bc.Iterator()
	for {
		block := bcIterator.Next()
		//统计未花费
		//获取block中的每个Transaction
		for i := len(block.Txs) - 1; i >= 0; i-- {
			unUTXOs = caculate(block.Txs[i], address, spentTxOutputs, unUTXOs)
		}
		//结束迭代
		hashInt := new(big.Int)
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(hashInt) == 0 {
			break
		}
	}
	return unUTXOs
}

func (bc *BlockChain) GetBalance(address string, txs []*Transaction) int64 {
	unUTXOs := bc.UnUTXOs(address, txs)
	fmt.Println(address, unUTXOs)
	var amount int64
	for _, utxo := range unUTXOs {
		amount = amount + utxo.Output.Value
	}
	return amount
}

//转账时查获可用的UTXO
func (bc *BlockChain) FindSpendableUTXOs(from string, amount int64, txs []*Transaction) (int64, map[string][]int) {
	/*
		1.获取所有的UTXO
		2.遍历UTXO

		返回值：map[hash](index)
	*/
	var balance int64
	utxos := bc.UnUTXOs(from, txs)
	fmt.Println(from, utxos)
	spendableUTXO := make(map[string][]int)
	for _, utxo := range utxos {
		balance += utxo.Output.Value
		hash := hex.EncodeToString(utxo.TxID)
		spendableUTXO[hash] = append(spendableUTXO[hash], utxo.Index)
		if balance >= amount {
			break
		}
	}
	if balance < amount {
		fmt.Printf("%s 余额不足。。总额：%d, 需要：%d\n", from, balance, amount)
		os.Exit(1)
	}
	return balance, spendableUTXO
}

func caculate(tx *Transaction, address string, spentTxOutputs map[string][]int, unUTXOs []*UTXO) []*UTXO {
	//先遍历TxInputs,表示花费。
	if !tx.IsCoinbaseTransaction() {
		for _, in := range tx.Vins {
			//如果解锁
			//base58解码
			fullPayloadHash := Base58Decode([]byte(address))
			//按位取出公钥作为地址
			pubKeyHash := fullPayloadHash[1 : len(fullPayloadHash)-addressChecksumLen]

			if in.UnLockWithAddress(pubKeyHash) {
				key := hex.EncodeToString(in.TxID)
				spentTxOutputs[key] = append(spentTxOutputs[key], in.Vout)
			}
		}
	}

	//遍历TxOutputs
outputs:
	for index, out := range tx.Vouts {
		if out.UnLockWithAddress(address) {
			//如果对应的花费容器中长度不为0
			if len(spentTxOutputs) != 0 {
				var isSpentUTXO bool

				for txID, indexArray := range spentTxOutputs {
					for _, i := range indexArray {
						if i == index && txID == hex.EncodeToString(tx.TxID) {
							isSpentUTXO = true
							continue outputs
						}
					}
				}
				if !isSpentUTXO {
					utxo := &UTXO{tx.TxID, index, out}
					unUTXOs = append(unUTXOs, utxo)
				}
			} else {
				utxo := &UTXO{tx.TxID, index, out}
				unUTXOs = append(unUTXOs, utxo)
			}
		}
	}
	return unUTXOs
}

//添加方法，签名一笔交易
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey, txs []*Transaction) {
	if tx.IsCoinbaseTransaction() {
		return
	}
	prevTxs := make(map[string]*Transaction)
	for _, vin := range tx.Vins {
		prevTx := bc.FindTransactionByTxID(vin.TxID, txs)
		prevTxs[hex.EncodeToString(prevTx.TxID)] = prevTx
	}
	tx.Sign(privKey, prevTxs)
}

//根据交易ID查找对应的Transaction
func (bc *BlockChain) FindTransactionByTxID(txID []byte, txs []*Transaction) *Transaction {
	iterator := bc.Iterator()
	//先遍历txs交易池中的交易
	for _, tx := range txs {
		if bytes.Compare(txID, tx.TxID) == 0 {
			return tx
		}
	}
	//再遍历区块链中的交易信息
	for {
		block := iterator.Next()
		for _, tx := range block.Txs {
			if bytes.Compare(txID, tx.TxID) == 0 {
				return tx
			}
		}
		//直到遍历到创世块的hashInt值为0，则停止。
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}
	return &Transaction{}
}

//验证数字签名
func (bc *BlockChain) VerifyTransaction(tx *Transaction, txs []*Transaction) bool {
	prevTXs := make(map[string]*Transaction)
	for _, vin := range tx.Vins {
		prevTX := bc.FindTransactionByTxID(vin.TxID, txs)
		prevTXs[hex.EncodeToString(prevTX.TxID)] = prevTX
	}
	return tx.Verify(prevTXs)
}

/*
查询未花费的output
[string] *TxOutputs
*/
func (bc *BlockChain) FindUnSpentOutputMap() map[string]*TxOutputs {
	iterator := bc.Iterator()

	//存储已经花费：[txID],txInput
	spentUTXOsMap := make(map[string][]*TXInput)
	//存储未花费
	unSpentOutputMaps := make(map[string]*TxOutputs)

	for {
		block := iterator.Next()

		for i := len(block.Txs) - 1; i >= 0; i-- {
			txOutputs := &TxOutputs{[]*UTXO{}}
			tx := block.Txs[i]

			if !tx.IsCoinbaseTransaction() {
				for _, txInput := range tx.Vins {
					key := hex.EncodeToString(txInput.TxID)
					spentUTXOsMap[key] = append(spentUTXOsMap[key], txInput)
				}
			}

			txID := hex.EncodeToString(tx.TxID)
		work:
			for index, out := range tx.Vouts {
				txInputs := spentUTXOsMap[txID]
				if len(txInputs) > 0 {
					var isSpent bool
					for _, input := range txInputs {
						inputPubKeyHash := PubKeyHash(input.PublicKey)
						if bytes.Compare(inputPubKeyHash, out.PubKeyHash) == 0 {
							if input.Vout == index {
								isSpent = true
								continue work
							}
						}
					}
					if isSpent == false {
						utxo := &UTXO{tx.TxID, index, out}
						txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
					}
				} else {
					utxo := &UTXO{tx.TxID, index, out}
					txOutputs.UTXOS = append(txOutputs.UTXOS, utxo)
				}
			}
			//设置
			unSpentOutputMaps[txID] = txOutputs
		}
		//停止迭代
		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return unSpentOutputMaps
}

//获取最新区块的高度
func (bc *BlockChain) GetBestHeight() int64 {
	block := bc.Iterator().Next()
	return block.Height
}

//获取所有区块的hash
func (bc *BlockChain) GetBlockHashes() [][]byte {
	blockIterator := bc.Iterator()
	var blockHashs [][]byte

	for {
		block := blockIterator.Next()

		blockHashs = append(blockHashs, block.Hash)

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)

		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}
	return blockHashs
}

//根据hash获取区块
func (bc *BlockChain) GetBlock(blockHash []byte) ([]byte, error) {
	var blockBytes []byte

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BLOCKTABLENAME))
		if b != nil {
			blockBytes = b.Get(blockHash)
		}
		return nil
	})

	return blockBytes, err
}

//添加区块到数据库
func (bc *BlockChain) AddBlock(block *Block) {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BLOCKTABLENAME))

		if b != nil {
			blockExist := b.Get(block.Hash)
			if blockExist != nil {
				//如果存在，不需要做任何过多的处理
				return nil
			}
			err := b.Put(block.Hash, block.Serilalize())
			if err != nil {
				log.Panic(err)
			}

			//最新的区块链的Hash
			blockHash := b.Get([]byte("1"))

			blockBytes := b.Get(blockHash)

			blockInDB := DeserializeBlock(blockBytes)

			if blockInDB.Height < block.Height {
				b.Put([]byte("1"), block.Hash)
				bc.Tip = block.Hash
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}
