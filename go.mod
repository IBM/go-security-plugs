module github.com/IBM/go-security-plugs

go 1.16

require (
	go.uber.org/zap v1.19.1
	knative.dev/serving v0.33.1-0.20220718185459-017b9d0393dd
)

replace github.com/IBM/go-security-plugs/plugs/rtgate => ./plugs/rtgate

replace github.com/IBM/go-security-plugs/pluginterfaces => ./pluginterfaces
