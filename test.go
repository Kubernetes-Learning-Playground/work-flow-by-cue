package main

import (
	"context"
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/tools/flow"
	"github.com/practice/workflow-practice/pkg/handlers"
	"log"
)

const (
	K8SFlowTpl = "pkg/flowtpls/k8sflow.cue"
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

	k8sflow := flow.New(&flow.Config{Root: cue.ParsePath(handlers.K8sTest1Root)}, cv, handlers.K8stest1Handler)
	check(k8sflow.Run(context.TODO()))
}
