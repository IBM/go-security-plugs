module github.com/IBM/go-security-plugs

go 1.16

require (
	go.uber.org/zap v1.19.1
	knative.dev/serving v0.33.1-0.20220725225524-63523f9d0e97
)

replace github.com/IBM/go-security-plugs/plugs/rtgate => ./plugs/rtgate

replace github.com/IBM/go-security-plugs/pluginterfaces => ./pluginterfaces

//replace knative.dev/serving => github.com/davidhadas/serving v0.27.1-0.20220727221439-b919980518c9
