package main

import (
	"fmt"

	"github.com/proudjiao/byte_douyin_project/config"
	"github.com/proudjiao/byte_douyin_project/router"
)

func main() {
	r := router.Init()
	err := r.Run(fmt.Sprintf(":%d", config.Global.Port)) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	if err != nil {
		return
	}
}
