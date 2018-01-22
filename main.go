package main

import (
	"fmt"
)

func main() {
	fmt.Println("Loading data...")
	HeroList, AllData:= GetDotaBuffData()
	terminalprocessor(HeroList,AllData)
}
