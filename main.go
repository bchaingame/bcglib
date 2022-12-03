package main

import (
	"bcglib/bcg"
	"time"
)

func main() {
	bcg.LogBlue(time.Now().Unix())
	bcg.LogBlue(time.Now().UnixMilli())
	bcg.LogBlue(time.Now().UnixMicro())
	bcg.LogBlue(time.Now().UnixNano())
}
