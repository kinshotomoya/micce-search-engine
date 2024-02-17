package main

import (
	"fmt"
	repository "synonym-data-generator/internal"
)

func main() {

	filePath := "./dict/synonym.csv"
	bblot, err := repository.NewBboltRepositpry()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = bblot.Write(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("create database done")

}
