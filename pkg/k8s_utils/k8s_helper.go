package k8s_utils

import (
	"bytes"
	"fmt"
	"io"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/util"
	"log"
)

/*
	操作通用k8s资源对象 如同kubectl
*/

// setDefaultNamespaceIfScopedAndNoneSet 设置namespace
func setDefaultNamespaceIfScopedAndNoneSet(u *unstructured.Unstructured, helper *resource.Helper) {
	namespace := u.GetNamespace()
	if helper.NamespaceScoped && namespace == "" {
		namespace = "default"
		u.SetNamespace(namespace)
	}
}

// newRestClient 初始化RestClient
// 可为每个资源对象创建各自的客户端
func newRestClient(restConfig *rest.Config, gv schema.GroupVersion) (rest.Interface, error) {
	restConfig.ContentConfig = resource.UnstructuredPlusDefaultContentConfig()
	restConfig.GroupVersion = &gv
	// 判断group是否存在
	if len(gv.Group) == 0 {
		restConfig.APIPath = "/api"
	} else {
		restConfig.APIPath = "/apis"
	}

	return rest.RESTClientFor(restConfig)
}

// K8sDelete k8s delete
func K8sDelete(json []byte, restConfig *rest.Config, mapper meta.RESTMapper) error {
	// 获取decoder对象
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(json), len(json))
	// 不断遍例
	for {
		var rawObj runtime.RawExtension
		err := decoder.Decode(&rawObj)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		// 得到gvk
		obj, gvk, err := syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			log.Fatal(err)
		}

		// 把obj变成map[string]interface{} -> unstructuredObj对象
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return nil
		}
		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}
		// 由gvk获取restMapping
		restMapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}
		// 获取客户端
		restClient, err := newRestClient(restConfig, gvk.GroupVersion())
		// 可以的操作实例
		helper := resource.NewHelper(restClient, restMapping)

		setDefaultNamespaceIfScopedAndNoneSet(unstructuredObj, helper)
		// 删除操作
		_, err = helper.Delete(unstructuredObj.GetNamespace(), unstructuredObj.GetName())
		if err != nil {
			log.Println(fmt.Sprintf("delete resource %s/%s fail:%s", unstructuredObj.GetNamespace(), unstructuredObj.GetName(), err.Error()))
		}

	}
	return nil
}

// K8sApply kubectl apply
func K8sApply(json []byte, restConfig *rest.Config, mapper meta.RESTMapper) ([]*resource.Info, error) {
	resList := make([]*resource.Info, 0)

	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(json), len(json))
	for {
		var rawObj runtime.RawExtension
		err := decoder.Decode(&rawObj)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return resList, err
			}
		}
		// 获取gvk
		obj, gvk, err := syaml.NewDecodingSerializer(unstructured.
			UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return resList, err
		}
		// obj 转成 map[string]interface{}
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return resList, err
		}
		unstructuredObj := &unstructured.Unstructured{Object: unstructuredMap}

		restMapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return resList, err
		}
		// 使用RestClient
		restClient, err := newRestClient(restConfig, gvk.GroupVersion())

		helper := resource.NewHelper(restClient, restMapping)

		setDefaultNamespaceIfScopedAndNoneSet(unstructuredObj, helper)

		objInfo := &resource.Info{
			Client:          restClient,
			Mapping:         restMapping,
			Namespace:       unstructuredObj.GetNamespace(),
			Name:            unstructuredObj.GetName(),
			Object:          unstructuredObj,
			ResourceVersion: restMapping.Resource.Version,
		}

		// kubectl 封装 的一个 patcher
		patcher, err := NewPatcher(objInfo, helper)
		if err != nil {
			return resList, err
		}

		// 获取更改的数据
		modified, err := util.GetModifiedConfiguration(objInfo.Object, true, unstructured.UnstructuredJSONScheme)
		if err != nil {
			return resList, err
		}

		if err := objInfo.Get(); err != nil {
			if !errors.IsNotFound(err) { //资源不存在
				return resList, err
			}

			// kubectl中的一些注解增加
			if err := util.CreateApplyAnnotation(objInfo.Object, unstructured.UnstructuredJSONScheme); err != nil {
				return resList, err
			}

			// 直接创建
			obj, err := helper.Create(objInfo.Namespace, true, objInfo.Object)
			if err != nil {
				return resList, err
			}
			objInfo.Refresh(obj, true)
		}

		_, patchedObject, err := patcher.Patch(objInfo.Object, modified, objInfo.Namespace, objInfo.Name)
		if err != nil {
			return resList, err
		}

		objInfo.Refresh(patchedObject, true)

		// ObjectInfo 放入列表
		resList = append(resList, objInfo)
	}
	return resList, nil
}
