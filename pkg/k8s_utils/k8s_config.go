package k8s_utils

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

var (
	K8sRestMapper *meta.RESTMapper
	K8sRestConfig *rest.Config
)

func init() {
	NewK8sConfig()
}

type K8sConfig struct{}

// NewK8sConfig 初始化
// RestConfig 与 RestMapper
func NewK8sConfig() *K8sConfig {
	cfg := &K8sConfig{}
	K8sRestMapper = cfg.RestMapper()
	K8sRestConfig = cfg.K8sRestConfig()
	return cfg
}

// K8sRestConfig 初始化RestConfig配置
func (kc *K8sConfig) K8sRestConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", "./resources/config")
	if err != nil {
		log.Fatal(err)
	}
	return config
}

// InitDynamicClient 初始化动态客户端
func (kc *K8sConfig) InitDynamicClient() dynamic.Interface {
	client, err := dynamic.NewForConfig(kc.K8sRestConfig())
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// RestMapper 获取所有资源对象 group-resource
// 初始化时先在内存保存，不需要重复从k8s中取
func (kc *K8sConfig) RestMapper() *meta.RESTMapper {
	gr, err := restmapper.GetAPIGroupResources(kc.InitClient().Discovery())
	if err != nil {
		log.Fatal(err)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(gr)
	return &mapper
}

// InitClient 初始化clientSet
func (kc *K8sConfig) InitClient() *kubernetes.Clientset {
	c, err := kubernetes.NewForConfig(kc.K8sRestConfig())
	if err != nil {
		log.Fatal(err)
	}
	return c
}
