package handlers

import (
	"cuelang.org/go/cue"
	"fmt"
	"github.com/practice/workflow-practice/pkg/utils"
	"k8s.io/cli-runtime/pkg/resource"
)


// 获取cue里面的值 ，必须传默认值   如果没有找到或出错，铁定会返回 默认值
func getField(value cue.Value, field string, defaultValue string) string {
	f := value.Value().LookupPath(cue.ParsePath(field))
	if !f.Exists() {
		return defaultValue
	}
	ret, err := f.String() //必须是string
	if err != nil {
		return defaultValue
	}
	return ret
}

// 就是 k8s apply
func apply(jsonBytes []byte) ([]*resource.Info, error) {
	return utils.K8sApply(jsonBytes, utils.K8sRestConfig, *utils.K8sRestMapper)
}

// 就是 k8s delete
func delete(jsonBytes []byte) error {
	return utils.K8sDelete(jsonBytes, utils.K8sRestConfig, *utils.K8sRestMapper)
}

func jsonField(v cue.Value, field string) ([]byte, error) {
	p := v.LookupPath(cue.ParsePath("body"))
	if !p.Exists() {
		return nil, fmt.Errorf("not found field:%s", field)
	}
	return p.MarshalJSON()
}
