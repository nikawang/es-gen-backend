package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	commonv1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/common/v1"
	esv1 "github.com/elastic/cloud-on-k8s/v2/pkg/apis/elasticsearch/v1"
	"github.com/labstack/echo/v4"
	"github.com/nikawang/es-gen-backend/model"
	"github.com/nikawang/es-gen-backend/sessionss"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// AccountController is a controller for managing user account.
type ESController interface {
	GetSession() sessionss.Session
	GetESList(c echo.Context) error
	GetES(c echo.Context) error
	UpdateES(c echo.Context) error
	CreateES(c echo.Context) error
	DeleteES(c echo.Context) error
}

type eSController struct {
	session sessionss.Session
}

func NewESController(s sessionss.Session) ESController {
	return &eSController{session: s}
}

func (controller *eSController) GetSession() sessionss.Session {
	return controller.session
}

func (controller *eSController) initK8SCfg(session sessionss.Session) *rest.Config {
	account := session.GetAccount()

	tlscfg := &rest.TLSClientConfig{
		Insecure: true,
	}

	// fmt.Printf("Username: %s\n", username)
	// fmt.Printf("Host: %s\n", account.host)
	// fmt.Printf("token: %s\n", account.token)
	cfg := &rest.Config{
		Host:            account.Host,
		BearerToken:     account.Token,
		TLSClientConfig: *tlscfg,
	}
	return cfg
}

func (controller *eSController) DeleteES(c echo.Context) error {
	fmt.Printf("DeleteES \n")

	cfg := controller.initK8SCfg(controller.GetSession())
	esv1.AddToScheme(scheme.Scheme)

	client, _ := k8sclient.New(cfg, k8sclient.Options{Scheme: scheme.Scheme})

	var es esv1.Elasticsearch

	eta := v1.ObjectMeta{
		Name:      c.Param("name"),
		Namespace: c.Param("namespace"),
	}
	es.ObjectMeta = eta

	if err := client.Delete(context.Background(), &es); err != nil {
		fmt.Println(err.Error())
	}

	return c.JSON(http.StatusOK, "OK")

}

func (controller *eSController) CreateES(c echo.Context) error {
	fmt.Printf("CreateES \n")

	var esDto model.ES
	// esDto := es.NewES()

	if err := c.Bind(&esDto); err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, esDto)
	}
	esLocalStr, _ := esDto.ToString()
	fmt.Printf("ES: \t%s", esLocalStr)

	var es esv1.Elasticsearch

	eta := v1.ObjectMeta{
		Name:      esDto.Name,
		Namespace: esDto.Namespace,
		Labels:    map[string]string{"esName": esDto.Name, "release": "es-generator"},
	}
	es.ObjectMeta = eta

	es.Spec.Version = esDto.Version

	for _, nodeSet := range esDto.NodeSets {
		withNodeSet(&es, nodeSet)
	}

	// fmt.Printf("es: \t %s", es)

	// fmt.Printf("ES:\nt %s\n", esLocalStr)
	cfg := controller.initK8SCfg(controller.GetSession())
	esv1.AddToScheme(scheme.Scheme)

	client, _ := k8sclient.New(cfg, k8sclient.Options{Scheme: scheme.Scheme})

	if err := client.Create(context.Background(), &es); err != nil {
		fmt.Println(err.Error())
	}

	return c.JSON(http.StatusOK, "OK")
}

func parseInt(i interface{}) (int, error) {
	str := fmt.Sprintf("%v", i)
	// if !ok {
	// 	return 0, errors.New("not string")
	// }
	return strconv.Atoi(str)
}
func withNodeSet(es *esv1.Elasticsearch, nsLocal model.NS) bool {
	var nodeSet esv1.NodeSet
	nodeSet.Config = &commonv1.Config{Data: map[string]interface{}{}}
	nodeSet.Config.Data["node.store.allow_mmap"] = true

	roles := strings.Split(nsLocal.NodeRole, "-")

	nodeSet.Config.Data["node.roles"] = roles
	nodeSet.Name = nsLocal.Name
	// nodeSet.Config.Data["noderoles"]

	count, _ := parseInt(nsLocal.Count)
	nodeSet.Count = int32(count)
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name: "elasticsearch-data",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(nsLocal.DiskSize),
				},
			},
			StorageClassName: &nsLocal.StorageClass,
		},
	}

	nodeSet.VolumeClaimTemplates = append(nodeSet.VolumeClaimTemplates, pvc)

	podAffinity := corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{{
				LabelSelector: &v1.LabelSelector{
					MatchExpressions: []v1.LabelSelectorRequirement{{
						Key:      "elasticsearch.k8s.elastic.co/cluster-name",
						Operator: "In",
						Values:   []string{es.Name},
					}},
				},
				TopologyKey: "kubernetes.io/hostname",
			}},
		},
	}

	nodeSet.PodTemplate.Spec.Affinity = &podAffinity

	es.Spec.NodeSets = append(es.Spec.NodeSets, nodeSet)
	return true
}

func (controller *eSController) UpdateES(c echo.Context) error {
	fmt.Printf("UpdateES \n")
	var esDto model.ES
	// esDto := es.NewES()

	if err := c.Bind(&esDto); err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusBadRequest, esDto)
	}

	esLocalStr, _ := esDto.ToString()
	fmt.Printf("ESName:\t%s\n", esDto.Name)
	fmt.Printf("ESNP:\t %s\n", esDto.Namespace)
	fmt.Printf("ES:\nt %s\n", esLocalStr)
	cfg := controller.initK8SCfg(controller.GetSession())
	esv1.AddToScheme(scheme.Scheme)

	client, _ := k8sclient.New(cfg, k8sclient.Options{Scheme: scheme.Scheme})

	var esMeta types.NamespacedName

	esMeta.Name = esDto.Name
	esMeta.Namespace = esDto.Namespace

	var es esv1.Elasticsearch

	if err := client.Get(context.Background(), esMeta, &es); err != nil {
		fmt.Println(err.Error())
	}

	// var *cfg rest.Config
	// dynamicClient, err := dynamic.NewForConfig(cfg)
	// if err != nil {
	// 	fmt.Printf("error creating dynamic client: %v\n", err)
	// 	// os.Exit(1)
	// }

	// gvr := schema.GroupVersionResource{
	// 	Group:    "elasticsearch.k8s.elastic.co",
	// 	Version:  "v1",
	// 	Resource: "elasticsearches",
	// }

	// data, err1 := dynamicClient.Resource(gvr).Namespace(esDto.Namespace).Get(context.Background(), esDto.Name, v1.GetOptions{})
	// if err1 != nil {
	// 	fmt.Printf("error getting es: %v\n", err1)
	// 	// os.Exit(1)
	// }

	// byteArr, err2 := data.MarshalJSON()

	// if err2 != nil {
	// 	fmt.Println(err2)
	// }

	// var es esv1.Elasticsearch
	// err3 := json.Unmarshal(byteArr, &es)

	// if err3 != nil {
	// 	fmt.Println(err3)
	// }
	// esUp := es.DeepCopy()

	for i := 0; i < len(esDto.NodeSets); i++ {
		nodeSetName1 := esDto.NodeSets[i].Name
		for j := 0; j < len(es.Spec.NodeSets); j++ {
			if es.Spec.NodeSets[j].Name == nodeSetName1 {
				fmt.Printf("Applying nodeSet Name: \t %s, Storage:\t %s \n", nodeSetName1, esDto.NodeSets[i].DiskSize)
				count, _ := parseInt(esDto.NodeSets[i].Count)
				es.Spec.NodeSets[j].Count = int32(count)
				fmt.Printf("count: \t%d", es.Spec.NodeSets[j].Count)
				resources := corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"storage": resource.MustParse(esDto.NodeSets[i].DiskSize),
					},
				}
				es.Spec.NodeSets[j].VolumeClaimTemplates[0].Spec.Resources = resources
			}
		}
	}
	// ecUp := es.DeepCopy()

	// clientset, err := kubernetes.NewForConfig(cfg)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	// return false
	// }

	// d, err3 := clientset.RESTClient().Get().AbsPath("/apis/elasticsearch.k8s.elastic.co/v1/namespaces/elastic/elasticsearches/quickstart").DoRaw(context.TODO())

	// if err3 != nil {
	// 	fmt.Println(err3.Error())
	// }

	// fmt.Printf("es: %s\n", d)

	// fmt.Printf("Commiting...\n")
	// url := fmt.Sprintf("/apis/elasticsearch.k8s.elastic.co/v1/namespaces/%s/elasticsearches/%s", esDto.Namespace, esDto.Name)
	// fmt.Printf("Commiting URL: \t %s...\n", url)
	// esUp1, err4 := clientset.RESTClient().Put().AbsPath(url).Body(esUp).VersionedParams(&v1.UpdateOptions{}, scheme.ParameterCodec).DoRaw(context.TODO())
	esUp1 := client.Update(context.Background(), &es)

	// if err4 != nil {
	// 	fmt.Println(err4.Error())
	// 	// return false
	// }
	// pods, err := clientset.CoreV1().Pods("").List(context.TODO(), v1.ListOptions{})

	// esUp, _ := dynamicClient.Resource(gvr).Namespace(esDto.Namespace).Update(context.TODO(), es, v1.UpdateOptions{})
	// nodeSetName := ""
	// es.Spec.NodeSets[nodeSetName].Count = 2

	return c.JSON(http.StatusOK, esUp1)

}

func (controller *eSController) GetES(c echo.Context) error {
	fmt.Printf("GetES %s\n", c.Param("name"))
	cfg := controller.initK8SCfg(controller.GetSession())

	esv1.AddToScheme(scheme.Scheme)

	client, _ := k8sclient.New(cfg, k8sclient.Options{Scheme: scheme.Scheme})

	var esMeta types.NamespacedName

	esMeta.Name = c.Param("name")
	esMeta.Namespace = c.Param("namespace")

	var es esv1.Elasticsearch

	if err := client.Get(context.Background(), esMeta, &es); err != nil {
		return err
	}
	// var *cfg rest.Config
	// dynamicClient, err := dynamic.NewForConfig(cfg)
	// if err != nil {
	// 	fmt.Printf("error creating dynamic client: %v\n", err)
	// 	// os.Exit(1)
	// }

	// gvr := schema.GroupVersionResource{
	// 	Group:    "elasticsearch.k8s.elastic.co",
	// 	Version:  "v1",
	// 	Resource: "elasticsearches",
	// }

	// data, err1 := dynamicClient.Resource(gvr).Namespace("elastic").Get(context.Background(), c.Param("name"), v1.GetOptions{})
	// if err1 != nil {
	// 	fmt.Printf("error getting es: %v\n", err1)
	// 	// os.Exit(1)
	// }

	// byteArr, err2 := data.MarshalJSON()

	// if err2 != nil {
	// 	fmt.Println(err2)
	// }

	// var es esv1.Elasticsearch
	// err3 := json.Unmarshal(byteArr, &es)

	// if err3 != nil {
	// 	fmt.Println(err3)
	// }

	// for _, es := range esList.Items {
	// 	fmt.Printf("ES Name : %s", es.Name)
	// 	fmt.Printf("ES Version : %s", es.Spec.Version)
	// 	fmt.Printf("ES Version : %s", es.Status.Health)
	// 	nodeSets := es.Spec.NodeSets
	// 	for _, nodeSet := range nodeSets {
	// 		fmt.Printf("NodeSet Name : %s\n", nodeSet.Name)
	// 		fmt.Printf("NodeSet Name : %d\n", nodeSet.Count)
	// 		fmt.Printf("NodeSet Name : %s\n", *nodeSet.VolumeClaimTemplates[0].Spec.StorageClassName)
	// 		fmt.Printf("NodeSet Name : %s\n", nodeSet.VolumeClaimTemplates[0].Spec.Resources.Requests.Storage().String())
	// 	}
	// }

	return c.JSON(http.StatusOK, es)
}

func (controller *eSController) GetESList(c echo.Context) error {
	fmt.Printf("GetESList\n")
	cfg := controller.initK8SCfg(controller.GetSession())
	// var *cfg rest.Config
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("error creating dynamic client: %v\n", err)
		// os.Exit(1)
	}

	gvr := schema.GroupVersionResource{
		Group:    "elasticsearch.k8s.elastic.co",
		Version:  "v1",
		Resource: "elasticsearches",
	}

	data, err1 := dynamicClient.Resource(gvr).List(context.Background(), v1.ListOptions{})
	if err1 != nil {
		fmt.Printf("error getting es: %v\n", err1)
		// os.Exit(1)
	}

	byteArr, err2 := data.MarshalJSON()

	if err2 != nil {
		fmt.Println(err2)
	}

	var esList esv1.ElasticsearchList
	err3 := json.Unmarshal(byteArr, &esList)

	if err3 != nil {
		fmt.Println(err3)
	}

	// dynamicClient.
	for _, es := range esList.Items {
		entryPoint, lbEntryPoint, password := getESEntryPoint(cfg, es.Name, es.Namespace)
		annotations := es.GetAnnotations()
		annotations["entryPoint"] = entryPoint
		annotations["lbEntryPoint"] = lbEntryPoint
		annotations["user"] = "elastic"
		annotations["password"] = password
		es.SetAnnotations(annotations)
	}
	return c.JSON(http.StatusOK, esList.Items)
}

func getESEntryPoint(cfg *rest.Config, esName string, esNS string) (string, string, string) {
	clientset, _ := kubernetes.NewForConfig(cfg)

	getOptions := v1.GetOptions{}
	fmt.Printf("esNamespace: %s, \t esName:%s \n", esNS, esName+"-es-http")
	svc, _ := clientset.CoreV1().Services(esNS).Get(context.Background(), esName+"-es-http", getOptions)
	svcEntryPoint := svc.Spec.ClusterIP + ":" + fmt.Sprint(svc.Spec.Ports[0].Port)
	var lbEntryPoint string
	lbEntryPoint = "_"
	if len(svc.Spec.LoadBalancerIP) > 0 {
		lbEntryPoint = svc.Spec.LoadBalancerIP + ":" + fmt.Sprint(svc.Spec.Ports[0].Port)
	}
	// lbEntryPoint := svc.Spec.LoadBalancerIP + ":" + fmt.Sprint(svc.Spec.Ports[0].Port)
	secret, _ := clientset.CoreV1().Secrets(esNS).Get(context.Background(), esName+"-es-elastic-user", getOptions)
	var password string
	for _, value := range secret.Data {
		// key is string, value is []byte
		fmt.Printf("pwdOrgin: %s\n", string(value))
		// pwdBytes, _ := (base64.StdEncoding.DecodeString(string(value)))
		password = string(value)
		// fmt.Printf("pwd:%s\n", password)
	}

	return svcEntryPoint, lbEntryPoint, password
	// if err != nil{
	//     log.Fatal(err)
	// }
	// for _, svc:=range svcs.Items{
	//     if strings.Contains(svc.Name, deployment){
	//         fmt.Fprintf(os.Stdout, "service name: %v\n", svc.Name)
	//         return &svc, nil
	//     }
	// }

}
