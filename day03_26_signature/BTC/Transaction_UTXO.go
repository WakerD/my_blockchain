package BTC

//创建一个结构体UTXO，用于表示所有未花费的
type UTXO struct {
	TxID   []byte    //当前Transaction的交易ID
	Index  int       //下标索引
	Output *TXOutput //要使用的未华为的Output
}
