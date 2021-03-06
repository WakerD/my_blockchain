package BTC

import (
	"fmt"
	"os"
)

//查询余额
func (cli *CLI) getBalance(address string, nodeID string) {
	fmt.Println("查询余额：", address)
	blockchain := GetBlockchainObject(nodeID)
	if blockchain == nil {
		fmt.Println("数据库不存在，无法查询。。")
		os.Exit(1)
	}

	defer blockchain.DB.Close()

	//balance := blockchain.GetBalance(address, []*Transaction{})
	utxoSet := &UTXOSet{blockchain}
	balance := utxoSet.GetBalance(address)
	fmt.Printf("%s,一共有%d个Token\n", address, balance)
}
