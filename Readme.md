# 步骤
* 下载k8s源码(code-generator,kubernetes等)代码
* 设置GOPATH
* 创建测试目录testcontroller,并构建以下目录树
```
wwh@wwh:~/kiongf/go/src/k8s1.9/src/testcontroller$ tree
.
└── pkg
    └── apis


```
* 在apis中创建api组目录树, 示例中api组是xfleet/v1,可根据实际项目定义api组 
```
wwh@wwh:~/kiongf/go/src/k8s1.9/src/testcontroller/pkg/apis$ tree
.
└── xfleet.com
    ├── register.go
    └── v1
        ├── register.go  //必须,否则无法注册api组
        └── types.go
        └── doc.go     //必须,否则无法通过deepcopy-gen生成deepcopy

```
* `xfleet.com/v1/doc.go`示例如下. 用于告知deepcopy-gen生成deepcopy文件. doc.go必须创建,而且不能删掉其中的注释.
```golang
// +k8s:deepcopy-gen=package,register

// Package v1 is the v1 version of the API.
// +groupName=xfleet.com
package v1
```
* `xfleet.com/register.go`示例如下. 用于注册api组
```
package xfleetcom

const (
  GroupName = "xfleet.com"
)
```
* `xfleet.com/v1/register.go`示例如下,用于注册api组版本
```
package v1

import (
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/apimachinery/pkg/runtime"
  "k8s.io/apimachinery/pkg/runtime/schema"

  examplecom "testcontroller/pkg/apis/xfleet.com"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: examplecom.GroupName, Version: "v1"}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
  return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
  // localSchemeBuilder and AddToScheme will stay in k8s.io/kubernetes.
  SchemeBuilder      runtime.SchemeBuilder
  localSchemeBuilder = &SchemeBuilder
  AddToScheme        = localSchemeBuilder.AddToScheme
)

func init() {
  // We only register manually written functions here. The registration of the
  // generated functions takes place in the generated files. The separation
  // makes the code compile even when the generated files are missing.
  localSchemeBuilder.Register(addKnownTypes)
}

// Adds the list of known types to api.Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
  scheme.AddKnownTypes(SchemeGroupVersion,
    &Foo{},
    &FooList{},
  )
  metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
  return nil 
}

```
* 在v1目录中type.go文件,描述第三方资源的定义.示例如下

```
package v1

import (
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Foo is a specification for a Foo resource
type Foo struct {
  metav1.TypeMeta   `json:",inline"`
  metav1.ObjectMeta `json:"metadata,omitempty"`

  Spec   FooSpec   `json:"spec"`
  Status FooStatus `json:"status"`
}

// FooSpec is the spec for a Foo resource
type FooSpec struct {
  DeploymentName string `json:"deploymentName"`
  Replicas       *int32 `json:"replicas"`
  Test           bool   `json:"test"`
}

// FooStatus is the status for a Foo resource
type FooStatus struct {
  AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FooList is a list of Foo resources
type FooList struct {
  metav1.TypeMeta `json:",inline"`
  metav1.ListMeta `json:"metadata"`

  Items []Foo `json:"items"`
}
  
```

* 返回到$GOPATH/src目录
* 执行以下命令
` k8s.io/code-generator/generate-groups.sh all testcontroller/pkg/client testcontroller/pkg/apis xfleet.com:v1
`
   * k8s.io/code-generator/generate-groups.sh: 执行的脚本
   * all: 表示构建的目标,all包括clientset, informer, doc等
   * testcontroller/pkg/client: 生成clientset,informer,doc存放的目录.这是相对路径, 根路径是是$GOPATH/src.可以通过在最后添加--output-base改变根路径.
   * testcontroller/pkg/apis: api组资源定义路径.里面可能包括api组.
   * xfleet:v1  : api组. 
* 命令执行后,生成依赖目录树
```
testcontroller/
├── main.go
└── pkg
    ├── apis
    │   └── xfleet.com
    │       ├── register.go
    │       └── v1
    │           ├── doc.go
    │           ├── register.go
    │           ├── types.go
    │           └── zz_generated.deepcopy.go
    └── client
        ├── clientset
        │   └── versioned
        │       ├── clientset.go
        │       ├── doc.go
        │       ├── fake
        │       │   ├── clientset_generated.go
        │       │   ├── doc.go
        │       │   └── register.go
        │       ├── scheme
        │       │   ├── doc.go
        │       │   └── register.go
        │       └── typed
        │           └── xfleet.com
        │               └── v1
        │                   ├── doc.go
        │                   ├── fake
        │                   │   ├── doc.go
        │                   │   ├── fake_foo.go
        │                   │   └── fake_xfleet.com_client.go
        │                   ├── foo.go
        │                   ├── generated_expansion.go
        │                   └── xfleet.com_client.go
        ├── informers
        │   └── externalversions
        │       ├── factory.go
        │       ├── generic.go
        │       ├── internalinterfaces
        │       │   └── factory_interfaces.go
        │       └── xfleet.com
        │           ├── interface.go
        │           └── v1
        │               ├── foo.go
        │               └── interface.go
        └── listers
            └── xfleet.com
                └── v1
                    ├── expansion_generated.go
                    └── foo.go


```

* 添加测试, 在testcontroller目录下创建main.go
```
package main

import (
  "flag"
  "fmt"

  "github.com/golang/glog"

  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/client-go/tools/clientcmd"

  examplecomclientset "testcontroller/pkg/client/clientset/versioned"
)

var (
  kuberconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
  master      = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
)

func main() {
  flag.Parse()

  cfg, err := clientcmd.BuildConfigFromFlags(*master, *kuberconfig)
  if err != nil {
    glog.Fatalf("Error building kubeconfig: %v", err)
  }

  exampleClient, err := examplecomclientset.NewForConfig(cfg)
  if err != nil {
    glog.Fatalf("Error building example clientset: %v", err)
  }

  list, err := exampleClient.XfleetV1().Foos("default").List(metav1.ListOptions{})
  if err != nil {
    glog.Fatalf("Error listing all foos: %v", err)
  }

  for _, db := range list.Items {
    fmt.Printf("database %s with user %q\n", db.Name, db.Spec.DeploymentName)
  }
}

```
* 这时候,还不能直接调用.因为资源定义还没有注册到k8s中.直接调用会报
```
 Error listing all foos: the server could not find the requested resource (get foos.xfleet.com)
```
* 这时候得创建一个crd,注册资源
```
package crd 

import (
  "reflect"

  apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
  apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
  apierrors "k8s.io/apimachinery/pkg/api/errors"
  meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  xfleetv1 "testcontroller/pkg/apis/xfleet.com/v1"
)

const (
  CRDPlural   string = "foos" //k8s如何描述资源,kubect get foos可以获取资源
  CRDGroup    string = "xfleet.com"
  CRDVersion  string = "v1"
  FullCRDName string = CRDPlural + "." + CRDGroup
)

// Create the CRD resource, ignore error if it already exists
func CreateCRD(clientset apiextcs.Interface) error {
  crd := &apiextv1beta1.CustomResourceDefinition{
    ObjectMeta: meta_v1.ObjectMeta{Name: FullCRDName},
    Spec: apiextv1beta1.CustomResourceDefinitionSpec{
      Group:   CRDGroup,
      Version: CRDVersion,
      Scope:   apiextv1beta1.NamespaceScoped,
      Names: apiextv1beta1.CustomResourceDefinitionNames{
        Plural: CRDPlural,
        Kind:   reflect.TypeOf(xfleetv1.Foo{}).Name(), //资源类型
      },  
    },  
  }

  _, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
  if err != nil && apierrors.IsAlreadyExists(err) {
    return nil 
  }
  return err 

  // Note the original apiextensions example adds logic to wait for creation and exception handling
}
```

* 在main函数中调用,创建crd
```

  apix "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
)

var (
  kuberconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
  master      = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
)

func main() {
  flag.Parse()

  cfg, err := clientcmd.BuildConfigFromFlags(*master, *kuberconfig)
  if err != nil {
    glog.Fatalf("Error building kubeconfig: %v", err)
  }

  clientset, err := apix.NewForConfig(cfg)
  if err != nil {
    panic(err)
  }
  err = crd.CreateCRD(clientset)
  if err != nil {
    panic(err)
  }

```
*  编译后执行
```
root@m-bang-192-168-9-203:~# kubectl get crd
NAME              AGE
foos.xfleet.com   11s
root@m-bang-192-168-9-203:~# kubectl get foo
No resources found.

```

# 参考
* https://github.com/yaronha/kube-crd 但注意,这个项目的作者并没有用code-generator生成clientset,而是通过代码写了一个
* https://thenewstack.io/extend-kubernetes-1-7-custom-resources