package main

import (
	"fmt"
	"time"

	"github.com/markaya/gonotes/cigbook"
)

func main() {
	fmt.Print("Start")

	cigbook.SyncCondExample()

	time.Sleep(2 * time.Second)
}
