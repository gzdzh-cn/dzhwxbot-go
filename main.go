package main

import (
	"github.com/gogf/gf/v2/os/gctx"

	"wxbot/internal/cmd"
	_ "wxbot/internal/packed"
)

func main() {

	cmd.Main.Run(gctx.New())

}
