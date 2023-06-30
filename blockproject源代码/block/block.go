package block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"jhblockchain/utils"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var MINING_DIFFICULT = 0x80000

const MINING_ACCOUNT_ADDRESS = "XYJ BLOCKCHAIN"
const MINING_REWARD = 5000
const MINING_TIMER_SEC = 10
const (
	//以下参数可以添加到启动参数
	BLOCKCHAIN_PORT_RANGE_START      = 5000
	BLOCKCHAIN_PORT_RANGE_END        = 5003
	NEIGHBOR_IP_RANGE_START          = 0
	NEIGHBOR_IP_RANGE_END            = 0
	BLOCKCHIN_NEIGHBOR_SYNC_TIME_SEC = 10
)

// 区块结构体
type Block struct {
	nonce        *big.Int
	previousHash [32]byte
	timestamp    int64
	transactions []*Transaction
	number       *big.Int
	difficulty   *big.Int
	hash         [32]byte
	txSize       uint16
}

func NewBlock(number *big.Int, nonce *big.Int, previousHash [32]byte, txs []*Transaction) *Block {
	b := new(Block)
	b.timestamp = time.Now().UnixNano()
	b.nonce = nonce
	b.previousHash = previousHash
	b.transactions = txs
	b.number = number
	b.txSize = uint16(len(txs))
	b.difficulty = big.NewInt(int64(MINING_DIFFICULT))
	b.hash = b.Hash()
	return b
}

func (b *Block) PreviousHash() [32]byte {
	return b.previousHash
}

func (b *Block) Nonce() *big.Int {
	return b.nonce
}

func (b *Block) Transactions() []*Transaction {
	return b.transactions
}

func (b *Block) Print() {
	log.Printf("%-15v:%30d\n", "timestamp", b.timestamp)
	log.Printf("%-15v:%30d\n", "nonce", b.nonce)
	log.Printf("%-15v:%30x\n", "previous_hash", b.previousHash)
	log.Printf("%-15v:%30x\n", "hash", b.hash)
	log.Printf("%-15v:%30d\n", "number", b.number)
	log.Printf("%-15v:%30d\n", "difficulty", b.difficulty)
	log.Printf("%-15v:%30d\n", "txSize", b.txSize)
	for _, t := range b.transactions {
		t.Print()
	}
}

type Blockchain struct {
	transactionPool   []*Transaction
	chain             []*Block
	blockchainAddress string
	port              uint16
	mux               sync.Mutex
	neighbors         []string
	muxNeighbors      sync.Mutex
}

// 新建一条链的第一个区块
// NewBlockchain(blockchainAddress string) *Blockchain
// 函数定义了一个创建区块链的方法，它接收一个字符串类型的参数 blockchainAddress，
// 它返回一个区块链类型的指针。在函数内部，它创建一个区块链对象并为其设置地址，
// 然后创建一个创世块并将其添加到区块链中，最后返回区块链对象。
func NewBlockchain(blockchainAddress string, port uint16) *Blockchain {
	bc := new(Blockchain)
	blocks, _ := ReadBlock()
	bc.chain = blocks
	bc.Print()
	if len(bc.chain) == 0 {
		b := &Block{}
		bc.CreateBlock(big.NewInt(0), big.NewInt(0), b.Hash()) //创世纪块
		bc.blockchainAddress = blockchainAddress
		bc.AddTransaction(MINING_ACCOUNT_ADDRESS, bc.blockchainAddress, big.NewInt(0), nil, nil)
	}
	bc.blockchainAddress = blockchainAddress
	bc.port = port
	return bc
}

// 将区块链信息写入txt文件
func (b *Block) WriteBlock() error {
	m, _ := b.MarshalJSON()
	// 打开文件，使用追加模式
	file, err := os.OpenFile("blockchain.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("无法打开文件：", err)
	}
	defer file.Close()

	// 写入字符串到文件
	_, err = file.WriteString(string(m[:]) + "\n")
	if err != nil {
		fmt.Println("无法写入文件：", err)
	}
	return err
}

// 从txt文件中读取区块信息
func ReadBlock() ([]*Block, error) {
	file, err := os.Open("blockchain.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	blocks := make([]*Block, 0)

	dec := json.NewDecoder(file)

	for dec.More() {
		var block *Block
		if err := dec.Decode(&block); err != nil {
			color.Red("无法加载区块")
			return nil, err
		}
		blocks = append(blocks, block)
	}

	return blocks, nil
}

func (bc *Blockchain) Chain() []*Block {
	return bc.chain
}

func (bc *Blockchain) Run() {

	bc.StartSyncNeighbors()
	bc.ResolveConflicts()
	bc.StartMining()
}

func (bc *Blockchain) SetNeighbors() {
	bc.neighbors = utils.FindNeighbors(
		utils.GetHost(), bc.port,
		NEIGHBOR_IP_RANGE_START, NEIGHBOR_IP_RANGE_END,
		BLOCKCHAIN_PORT_RANGE_START, BLOCKCHAIN_PORT_RANGE_END)

	color.Blue("邻居节点：%v", bc.neighbors)
}

func (bc *Blockchain) SyncNeighbors() {
	bc.muxNeighbors.Lock()
	defer bc.muxNeighbors.Unlock()
	bc.SetNeighbors()
}

func (bc *Blockchain) StartSyncNeighbors() {
	bc.SyncNeighbors()
	_ = time.AfterFunc(time.Second*BLOCKCHIN_NEIGHBOR_SYNC_TIME_SEC, bc.StartSyncNeighbors)
}

func (bc *Blockchain) TransactionPool() []*Transaction {
	return bc.transactionPool
}

func (bc *Blockchain) ClearTransactionPool() {
	bc.transactionPool = bc.transactionPool[:0]
	color.Magenta("%x", len(bc.transactionPool))
	blocks, _ := ReadBlock()
	bc.chain = blocks
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*Block `json:"chain"`
	}{
		Blocks: bc.chain,
	})
}

func (bc *Blockchain) UnmarshalJSON(data []byte) error {
	v := &struct {
		Blocks *[]*Block `json:"chain"`
	}{
		Blocks: &bc.chain,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}

// (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block
//  函数是在区块链上创建新的区块，它接收两个参数：一个int类型的nonce和一个字节数组类型的 previousHash，
//  返回一个区块类型的指针。在函数内部，它使用传入的参数来创建一个新的区块，
//  然后将该区块添加到区块链的链上，并清空交易池。

func (bc *Blockchain) CreateBlock(number *big.Int, nonce *big.Int, previousHash [32]byte) *Block {
	b := NewBlock(number, nonce, previousHash, bc.transactionPool)

	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*Transaction{}

	err := b.WriteBlock()
	if err != nil {
		log.Fatal("写入区块失败", err)
	}

	// 删除其他节点的交易
	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/transactions", n)
		client := &http.Client{}
		req, _ := http.NewRequest("DELETE", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}

	return b
}

// 根据区块号查询区块
func (bc *Blockchain) GetBlockByNumber(blockid uint64) (*Block, error) {
	for i, block := range bc.chain {
		if uint64(i) == blockid {
			color.Green("%s BLOCK %d %s\n", strings.Repeat("=", 25), blockid, strings.Repeat("=", 25))
			block.Print()
			return block, nil
		}
	}
	return nil, nil
}

// 根据哈希查询区块
func (bc *Blockchain) GetBlockByHash(hash [32]byte) (*Block, error) {
	for _, block := range bc.chain {
		if block.hash == hash {
			color.Green("%s BLOCK %d %s\n", strings.Repeat("=", 25), block.number, strings.Repeat("=", 25))
			block.Print()
			return block, nil
		}
	}
	return nil, nil
}

func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		color.Green("%s BLOCK %d %s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		block.Print()
	}
	color.Yellow("%s\n\n\n", strings.Repeat("*", 50))
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256([]byte(m))
}

func (b *Block) MarshalJSON() ([]byte, error) {

	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        *big.Int       `json:"nonce"`
		PreviousHash string         `json:"previous_hash"`
		Transactions []*Transaction `json:"transactions"`
		Hash         string         `json:"hash"`
		Number       *big.Int       `json:"number"`
		Difficulty   *big.Int       `json:"difficulty"`
		TxSize       uint16         `json:"txSize"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		PreviousHash: fmt.Sprintf("%x", b.previousHash),
		Transactions: b.transactions,
		Hash:         fmt.Sprintf("%x", b.hash),
		Number:       b.number,
		Difficulty:   b.difficulty,
		TxSize:       b.txSize,
	})
}

func (b *Block) UnmarshalJSON(data []byte) error {
	var previousHash string
	var hash string
	var nonce int64
	var number int64
	var difficulty int64
	v := &struct {
		Timestamp    *int64          `json:"timestamp"`
		Nonce        *int64          `json:"nonce"`
		PreviousHash *string         `json:"previous_hash"`
		Transactions *[]*Transaction `json:"transactions"`
		Hash         *string         `json:"hash"`
		Number       *int64          `json:"number"`
		Difficulty   *int64          `json:"difficulty"`
		TxSize       *uint16         `json:"txSize"`
	}{
		Timestamp:    &b.timestamp,
		Nonce:        &nonce,
		PreviousHash: &previousHash,
		Transactions: &b.transactions,
		Hash:         &hash,
		Number:       &number,
		Difficulty:   &difficulty,
		TxSize:       &b.txSize,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	b.nonce = big.NewInt(nonce)
	b.difficulty = big.NewInt(difficulty)
	b.number = big.NewInt(number)

	ph, _ := hex.DecodeString(*v.PreviousHash)
	copy(b.previousHash[:], ph[:32])

	h, _ := hex.DecodeString(*v.Hash)
	copy(b.hash[:], h[:32])
	return nil
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) AddTransaction(
	sender string,
	recipient string,
	value *big.Int,
	senderPublicKey *ecdsa.PublicKey,
	s *utils.Signature) bool {
	t := NewTransaction(sender, recipient, value)

	//如果是挖矿得到的奖励交易，不验证
	if sender == MINING_ACCOUNT_ADDRESS {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}

	// 判断有没有足够的余额
	log.Printf("transaction.go sender:%s  account=%d", sender, bc.CalculateTotalAmount(sender))
	if bc.CalculateTotalAmount(sender).Cmp(value) < 0 {
		color.Red("ERROR: %s ，你的钱包里没有足够的钱", sender)
		return false
	}

	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {

		bc.transactionPool = append(bc.transactionPool, t)
		return true
	} else {
		log.Println("ERROR: 验证交易")
	}
	return false

}

func (bc *Blockchain) CreateTransaction(sender string, recipient string, value *big.Int,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	isTransacted := bc.AddTransaction(sender, recipient, value, senderPublicKey, s)

	if isTransacted {
		for _, n := range bc.neighbors {
			publicKeyStr := fmt.Sprintf("%064x%064x", senderPublicKey.X.Bytes(),
				senderPublicKey.Y.Bytes())
			signatureStr := s.String()
			bt := &TransactionRequest{
				&sender, &recipient, &publicKeyStr, value, &signatureStr}
			m, _ := json.Marshal(bt)
			buf := bytes.NewBuffer(m)
			endpoint := fmt.Sprintf("http://%s/transactions", n)
			client := &http.Client{}
			req, _ := http.NewRequest("PUT", endpoint, buf)
			resp, _ := client.Do(req)
			log.Printf("   **  **  **  CreateTransaction : %v", resp)
		}
	}

	return isTransacted
}

func (bc *Blockchain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.transactionPool {
		transactions = append(transactions,
			NewTransaction(t.senderAddress,
				t.receiveAddress,
				t.value))
	}
	return transactions
}

func bytesToBigInt(b [32]byte) *big.Int {
	bytes := b[:]
	result := new(big.Int).SetBytes(bytes)
	return result
}

func (bc *Blockchain) ValidProof(nonce *big.Int,
	previousHash [32]byte,
	transactions []*Transaction,
	difficulty *big.Int,
) bool {
	bigi_2 := big.NewInt(2)
	bigi_256 := big.NewInt(256)
	bigi_diff := difficulty
	target := new(big.Int).Exp(bigi_2, bigi_256, nil)
	target = new(big.Int).Div(target, bigi_diff)
	tmpBlock := Block{nonce: nonce, previousHash: previousHash, transactions: transactions, timestamp: 0}
	result := bytesToBigInt(tmpBlock.Hash())
	return target.Cmp(result) > 0
}

func (bc *Blockchain) getBlockSendTime(bum int) int {
	if bum == 0 {
		return 0
	}
	return int(bc.chain[bum].timestamp - bc.chain[bum-1].timestamp)
}

func (bc *Blockchain) ProofOfWork() *big.Int {
	transactions := bc.CopyTransactionPool() //选择交易？控制交易数量？
	previousHash := bc.LastBlock().Hash()
	nonce := big.NewInt(0)
	begin := time.Now()
	if bc.getBlockSendTime(len(bc.chain)-1) < 3e+9 {
		MINING_DIFFICULT += 32

	} else {
		if MINING_DIFFICULT >= 13000 {
			MINING_DIFFICULT -= 32
		}
	}
	for !bc.ValidProof(nonce, previousHash, transactions, big.NewInt(int64(MINING_DIFFICULT))) {
		one := big.NewInt(1)
		nonce.Add(nonce, one)
	}
	end := time.Now()

	log.Printf("POW spend Time:%f Second", end.Sub(begin).Seconds())
	log.Printf("POW spend Time:%s", end.Sub(begin))

	return nonce
}

// 将交易池的交易打包
func (bc *Blockchain) Mining() bool {
	bc.mux.Lock()

	defer bc.mux.Unlock()

	// 此处判断交易池是否有交易，你可以不判断，打包无交易区块
	if len(bc.transactionPool) == 0 {
		// color.Magenta("打包失败")
		return false
	}

	mining_reward, _ := big.NewFloat(MINING_REWARD).Int(nil)
	bc.AddTransaction(MINING_ACCOUNT_ADDRESS, bc.blockchainAddress, mining_reward, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().hash
	number := len(bc.chain)
	bc.CreateBlock(big.NewInt(int64(number)), nonce, previousHash)
	log.Println("action=mining, status=success")

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/consensus", n)
		client := &http.Client{}
		req, _ := http.NewRequest("PUT", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}

	color.Magenta("打包成功")
	return true
}

func (bc *Blockchain) CalculateTotalAmount(accountAddress string) *big.Int {
	var totalAmount *big.Int = big.NewInt(0)
	for _, _chain := range bc.chain {
		for _, _tx := range _chain.transactions {
			if accountAddress == _tx.receiveAddress {
				totalAmount.Add(totalAmount, _tx.value)
			}
			if accountAddress == _tx.senderAddress {
				totalAmount.Sub(totalAmount, _tx.value)
			}
		}
	}
	return totalAmount
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	// 使用time.AfterFunc函数创建了一个定时器，它在指定的时间间隔后执行bc.StartMining函数（自己调用自己）。
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
	color.Yellow("minetime: %v\n", time.Now())
}

type AmountResponse struct {
	Amount *big.Int `json:"amount"`
}

func (ar *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount *big.Int `json:"amount"`
	}{
		Amount: ar.Amount,
	})
}

type Transaction struct {
	senderAddress  string
	receiveAddress string
	value          *big.Int
	hash           [32]byte
}

func NewTransaction(sender string, receive string, value *big.Int) *Transaction {
	t := new(Transaction)
	t.senderAddress = sender
	t.receiveAddress = receive
	t.value = value
	t.hash = t.Hash()
	return t
}

func (t *Transaction) Hash() [32]byte {
	m, _ := json.Marshal(t)
	return sha256.Sum256([]byte(m))
}

func (bc *Blockchain) VerifyTransactionSignature(
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
	h := t.hash
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
}

func (bc *Blockchain) GetTransactionByHash(hash [32]byte) *Transaction {
	for _, block := range bc.chain {
		for _, transaction := range block.transactions {
			if transaction.hash == hash {
				return transaction
			}
		}
	}
	return nil
}

func (bc *Blockchain) GetTransactions() []*Transaction {
	var transactions []*Transaction
	for _, block := range bc.chain {
		transactions = append(transactions, block.transactions...)
	}
	return transactions
}

func (t *Transaction) Print() {
	color.Red("%s\n", strings.Repeat("~", 30))
	color.Cyan("发送地址             %s\n", t.senderAddress)
	color.Cyan("接受地址             %s\n", t.receiveAddress)
	color.Cyan("金额                 %d\n", t.value)

}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string   `json:"sender_blockchain_address"`
		Recipient string   `json:"recipient_blockchain_address"`
		Value     *big.Int `json:"value"`
		Hash      string   `json:"hash"`
	}{
		Sender:    t.senderAddress,
		Recipient: t.receiveAddress,
		Value:     t.value,
		Hash:      fmt.Sprintf("%x", t.hash),
	})
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	var hash string
	var value int64
	v := &struct {
		Sender    *string `json:"sender_blockchain_address"`
		Recipient *string `json:"recipient_blockchain_address"`
		Value     *int64  `json:"value"`
		Hash      *string `json:"hash"`
	}{
		Sender:    &t.senderAddress,
		Recipient: &t.receiveAddress,
		Value:     &value,
		Hash:      &hash,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	h, _ := hex.DecodeString(*v.Hash)
	copy(t.hash[:], h[:32])

	t.value = big.NewInt(value)
	return nil
}

func (bc *Blockchain) ValidChain(chain []*Block) bool {
	preBlock := chain[0]
	currentIndex := 1
	for currentIndex < len(chain) {
		b := chain[currentIndex]
		if b.previousHash != preBlock.Hash() {
			return false
		}

		if !bc.ValidProof(b.Nonce(), b.PreviousHash(), b.Transactions(), big.NewInt(int64(MINING_DIFFICULT))) {
			return false
		}

		preBlock = b
		currentIndex += 1
	}
	return true
}

func (bc *Blockchain) ResolveConflicts() bool {
	var longestChain []*Block = nil
	maxLength := len(bc.chain)

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/chain", n)
		resp, err := http.Get(endpoint)
		if err != nil {
			color.Red("                 错误 ：ResolveConflicts GET请求")
			return false
		} else {
			color.Green("                正确 ：ResolveConflicts  GET请求")
		}
		if resp.StatusCode == 200 {
			var bcResp Blockchain
			decoder := json.NewDecoder(resp.Body)
			err1 := decoder.Decode(&bcResp)

			if err1 != nil {
				color.Red("                 错误 ：ResolveConflicts Decode")
				return false
			} else {
				color.Green("                正确 ：ResolveConflicts  Decode")
			}

			chain := bcResp.Chain()
			color.Cyan("   ResolveConflicts   chain len:%d ", len(chain))
			if len(chain) > maxLength && bc.ValidChain(chain) {
				maxLength = len(chain)
				longestChain = chain
			}
		}
	}

	color.Cyan("   ResolveConflicts   longestChain len:%d ", len(longestChain))

	if longestChain != nil {
		bc.chain = longestChain
		log.Printf("Resovle confilicts replaced")
		return true
	}
	log.Printf("Resovle conflicts not replaced")
	return false
}

type TransactionRequest struct {
	SenderBlockchainAddress    *string  `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string  `json:"recipient_blockchain_address"`
	SenderPublicKey            *string  `json:"sender_public_key"`
	Value                      *big.Int `json:"value"`
	Signature                  *string  `json:"signature"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Value == nil ||
		tr.Signature == nil {
		return false
	}
	return true
}
