module github.com/IBM/go-security-plugs

go 1.16

require (
	github.com/kelseyhightower/envconfig v1.4.0
	go.uber.org/zap v1.19.1
	k8s.io/apimachinery v0.24.3 // indirect
	knative.dev/pkg v0.0.0-20220715183228-f1f36a2c977e // indirect
	knative.dev/serving v0.0.0-00010101000000-000000000000
)

replace github.com/IBM/go-security-plugs/plugs/rtgate => ./plugs/rtgate

replace github.com/IBM/go-security-plugs/pluginterfaces => ./pluginterfaces

replace knative.dev/serving => ../../davidhadas/serving

replace github.com/IBM/workload-security-guard => ../workload-security-guard
