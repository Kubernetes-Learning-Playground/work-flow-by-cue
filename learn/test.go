package main

import (
	"context"
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/tools/flow"
	"log"
)

const (
	K8SFlowTpl = "pkg/work_flow_template/k8sflow.cue"
)

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	inst := load.Instances([]string{K8SFlowTpl}, nil)[0]

	cc := cuecontext.New()
	cv := cc.BuildInstance(inst)

	k8sflow := flow.New(&flow.Config{Root: cue.ParsePath(K8sTest1Root)}, cv, K8stest1Handler)
	check(k8sflow.Run(context.TODO()))
}
