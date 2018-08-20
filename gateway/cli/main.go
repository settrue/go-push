package main

import "github.com/settrue/go-push/gateway"

func main () {

	gateway.InitConf()

	gateway.InitMaster()

	gateway.InitMerger()

	gateway.InitService()
}