package main

import (
	"fmt"
	"jhblockchain/utils"
)

func main() {
	hostip := utils.GetHost()
	fmt.Println("HOST IP:", hostip)
	fmt.Println(utils.FindNeighbors(hostip, 5000, 0, 3, 5000, 5003))
}
