package handlers

import (
	"cuelang.org/go/cue"
	"fmt"
	"github.com/practice/workflow-practice/pkg/k8s_utils"
	"k8s.io/cli-runtime/pkg/resource"
)

// getField 获取cue里面的值 ，必须传默认值，如果没有找到或出错，会返回默认值
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

// jsonField 会获取到 template 字段下的 json对象
func jsonField(v cue.Value, field string) ([]byte, error) {
	p := v.LookupPath(cue.ParsePath("template"))
	if !p.Exists() {
		return nil, fmt.Errorf("not found field:%s", field)
	}
	return p.MarshalJSON()
}

// 传入json bytes 使用 k8s apply
func apply(jsonBytes []byte) ([]*resource.Info, error) {
	return k8s_utils.K8sApply(jsonBytes, k8s_utils.K8sRestConfig, *k8s_utils.K8sRestMapper)
}

// 传入json bytes 使用 k8s delete
func delete(jsonBytes []byte) error {
	return k8s_utils.K8sDelete(jsonBytes, k8s_utils.K8sRestConfig, *k8s_utils.K8sRestMapper)
}
