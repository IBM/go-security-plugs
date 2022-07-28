module github.com/IBM/go-security-plugs

go 1.16

require (
	go.uber.org/zap v1.19.1
	knative.dev/serving v0.0.0-00010101000000-000000000000
)

replace github.com/IBM/go-security-plugs/plugs/rtgate => ./plugs/rtgate

replace github.com/IBM/go-security-plugs/pluginterfaces => ./pluginterfaces

replace knative.dev/serving => github.com/davidhadas/serving v0.27.1-0.20220727221439-b919980518c9
