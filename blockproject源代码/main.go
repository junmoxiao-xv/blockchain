package main

import (
	"fmt"
	"jhblockchain/block"
	"jhblockchain/wallet"
	"log"
	"math/big"

	"github.com/fatih/color"
)

func init() {

	color.Green("     ██╗ ██████╗ ██╗  ██╗███╗   ██╗██╗  ██╗ █████╗ ██╗")
	color.Green("     ██║██╔═══██╗██║  ██║████╗  ██║██║  ██║██╔══██╗██║")
	color.Green("     ██║██║   ██║███████║██╔██╗ ██║███████║███████║██║")
	color.Green("██   ██║██║   ██║██╔══██║██║╚██╗██║██╔══██║██╔══██║██║")
	color.Green("╚█████╔╝╚██████╔╝██║  ██║██║ ╚████║██║  ██║██║  ██║██║")
	color.Green("╚════╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝")

	color.Red("██████╗ ██╗      ██████╗  ██████╗██╗  ██╗ ██████╗██╗  ██╗ █████╗ ██╗███╗   ██╗")
	color.Red("██╔══██╗██║     ██╔═══██╗██╔════╝██║ ██╔╝██╔════╝██║  ██║██╔══██╗██║████╗  ██║")
	color.Red("██████╔╝██║     ██║   ██║██║     █████╔╝ ██║     ███████║███████║██║██╔██╗ ██║")
	color.Red("██╔══██╗██║     ██║   ██║██║     ██╔═██╗ ██║     ██╔══██║██╔══██║██║██║╚██╗██║")
	color.Red("██████╔╝███████╗╚██████╔╝╚██████╗██║  ██╗╚██████╗██║  ██║██║  ██║██║██║ ╚████║")
	color.Red("╚═════╝ ╚══════╝ ╚═════╝  ╚═════╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝")

	log.SetPrefix("Blockchain: ")
}

func main() {

	wallet_swk := wallet.NewWallet()     //孙悟空
	wallet_zbj := wallet.NewWallet()     //猪八戒
	wallet_johnhai := wallet.NewWallet() //矿工

	fmt.Printf("孙悟空的account:%s\n", wallet_swk.BlockchainAddress())
	fmt.Printf("猪八戒的account:%s\n", wallet_zbj.BlockchainAddress())
	fmt.Printf("矿工  的account:%s\n", wallet_johnhai.BlockchainAddress())

	blockchain := block.NewBlockchain(wallet_johnhai.BlockchainAddress(), 5000)
	blockchain.Mining()
	blockchain.Print()
	//钱包 提交一笔交易
	t := wallet.NewTransaction(
		wallet_johnhai.PrivateKey(),
		wallet_johnhai.PublicKey(),
		wallet_johnhai.BlockchainAddress(),
		wallet_zbj.BlockchainAddress(),
		8)

	//区块链 打包交易
	isAdded := blockchain.AddTransaction(
		wallet_johnhai.BlockchainAddress(),
		wallet_zbj.BlockchainAddress(),
		big.NewInt(8),
		wallet_johnhai.PublicKey(),
		t.GenerateSignature())

	fmt.Println("这笔交易验证通过吗? ", isAdded)

	t2 := wallet.NewTransaction(
		wallet_swk.PrivateKey(),
		wallet_swk.PublicKey(),
		wallet_swk.BlockchainAddress(),
		wallet_zbj.BlockchainAddress(),
		80)

	//区块链 打包交易
	isAdded = blockchain.AddTransaction(
		wallet_swk.BlockchainAddress(),
		wallet_zbj.BlockchainAddress(),
		big.NewInt(80),
		wallet_swk.PublicKey(),
		t2.GenerateSignature())

	fmt.Println("这笔交易验证通过吗? ", isAdded)

	blockchain.Mining()
	blockchain.Print()

	fmt.Printf("孙悟空 %d\n", blockchain.CalculateTotalAmount(wallet_swk.BlockchainAddress()))
	fmt.Printf("猪八戒 %d\n", blockchain.CalculateTotalAmount(wallet_zbj.BlockchainAddress()))
	fmt.Printf("矿工   %d\n", blockchain.CalculateTotalAmount(wallet_johnhai.BlockchainAddress()))

	// w := wallet.NewWallet()
	// fmt.Println("PrivateKeyStr:", w.PrivateKeyStr())
	// fmt.Println("PublicKeyStr", w.PublicKeyStr())
	// fmt.Println("BlockchainAddress==>", w.BlockchainAddress())

	// t := wallet.NewTransaction(w.PrivateKey(), w.PublicKey(), w.BlockchainAddress(), "B", 1.0)
	// fmt.Printf("交易t:%+v\n", t)
	// fmt.Printf("signature签名: %s\n", t.GenerateSignature())

	// fmt.Println("PrivateKey==>\n", w.PrivateKey())
	// fmt.Println("PublicKey==>\n", w.PublicKey())

	// wallet.FromPriKeyToPubKey("7a6026c4bc5f70169cfc16f60cb9283a011a4e452a67151164c992fe50ac3fc4")

	// wallet.PrintEthereumAccount()

	// BigintDemo()
	// minerAddress := "江海"

	// blockChain := block.NewBlockchain(minerAddress)
	// blockChain.AddTransaction("孙悟空", "猪八戒", 20)
	// blockChain.Mining()
	// blockChain.Print()

	// blockChain.AddTransaction("孙悟空", "唐僧", 10)
	// blockChain.AddTransaction("唐僧", "猪八戒", 5)
	// blockChain.Mining()
	// blockChain.Print()

	// account0 := "孙悟空"
	// fmt.Printf("账户:%s 的余额是：%d\n", account0, blockChain.CalculateToatlAmount(account0))
	// account1 := "唐僧"
	// fmt.Printf("账户:%s 的余额是：%d\n", account1, blockChain.CalculateToatlAmount(account1))
	// account2 := "猪八戒"
	// fmt.Printf("账户:%s 的余额是：%d\n", account2, blockChain.CalculateToatlAmount(account2))

}
