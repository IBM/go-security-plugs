package main

import (
	"github.com/IBM/go-security-plugs/qpsecurity"
	"knative.dev/serving/pkg/queue/sharedmain" // use go get knative.dev/serving/sharedmain@QPShimMain
)

func main() {
	qOpt := qpsecurity.NewQPSecurityPlugs()
	defer qOpt.Shutdown()

	sharedmain.Main(qOpt.QPTransport)
}
