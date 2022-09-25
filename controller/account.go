package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/nikawang/es-gen-backend/model"
	"github.com/nikawang/es-gen-backend/sessionss"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// AccountController is a controller for managing user account.
type AccountController interface {
	GetLoginStatus(c echo.Context) error
	GetLoginAccount(c echo.Context) error
	Login(c echo.Context) error
	GetSession() sessionss.Session
	GetClientSet() kubernetes.Clientset
	Logout(c echo.Context) error
}

type accountController struct {
	session   sessionss.Session
	clientSet kubernetes.Clientset
}

func NewAccountController(s sessionss.Session) AccountController {
	return &accountController{session: s}
}

func (controller *accountController) GetLoginStatus(c echo.Context) error {
	return c.JSON(http.StatusOK, true)
}

func (controller *accountController) Login(c echo.Context) error {
	var account model.Account
	if err := c.Bind(&account); err != nil {
		return c.JSON(http.StatusBadRequest, account)
	}
	username, parsed := controller.ParseToken(account.Token)

	if !parsed {
		return c.NoContent(http.StatusUnauthorized)
	}

	fmt.Printf("Username: %s\n", username)
	sess := controller.GetSession()

	// if accCache := sess.GetAccount(); accCache != nil {
	// 	return c.JSON(http.StatusOK, accCache)
	// }

	authenticate := controller.Authenticate(account.Host, account.Token)
	if authenticate {
		account.Name = username
		// account.Cfg = cfg
		_ = sess.SetAccount(&account)
		_ = sess.Save()
		// controller.clientSet = *clientSet
		fmt.Println("Authorized")
		return c.JSON(http.StatusOK, account)
	}
	fmt.Println("Not Authorized")
	return c.NoContent(http.StatusUnauthorized)
}

func (controller *accountController) GetSession() sessionss.Session {
	return controller.session
}

func (controller *accountController) GetClientSet() kubernetes.Clientset {
	return controller.clientSet
}

func (controller *accountController) Authenticate(host string, token string) bool {
	tlscfg := &rest.TLSClientConfig{
		Insecure: true,
	}

	// fmt.Printf("Username: %s\n", username)
	fmt.Printf("Host: %s\n", host)
	fmt.Printf("token: %s\n", token)
	cfg := &rest.Config{
		Host:            host,
		BearerToken:     token,
		TLSClientConfig: *tlscfg,
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Println(err.Error())
		return false
		// panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	// d, err := clientset.RESTClient().Get().AbsPath("/apis/elasticsearch.k8s.elastic.co/v1/namespaces/elastic/elasticsearches").DoRaw(context.TODO())

	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("es: %s\n", d)

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

	// data, err1 := dynamicClient.Resource(gvr).Namespace("elastic").List(context.Background(), v1.ListOptions{})
	// if err1 != nil {
	// 	fmt.Printf("error getting es: %v\n", err1)
	// 	// os.Exit(1)
	// }

	// byteArr, err2 := data.MarshalJSON()

	// if err2 != nil {
	// 	fmt.Println(err2)
	// }

	// var esList esv1.ElasticsearchList
	// err3 := json.Unmarshal(byteArr, &esList)

	// if err3 != nil {
	// 	fmt.Println(err3)
	// }

	// for _, es := range esList.Items {
	// 	fmt.Printf("ES Name : %s", es.Name)
	// 	fmt.Printf("ES Version : %s", es.Spec.Version)

	// 	nodeSets := es.Spec.NodeSets
	// 	for _, nodeSet := range nodeSets {
	// 		fmt.Printf("NodeSet Name : %s\n", nodeSet.Name)
	// 		fmt.Printf("NodeSet Name : %d\n", nodeSet.Count)
	// 		fmt.Printf("NodeSet Name : %s\n", *nodeSet.VolumeClaimTemplates[0].Spec.StorageClassName)
	// 		fmt.Printf("NodeSet Name : %s\n", nodeSet.VolumeClaimTemplates[0].Spec.Resources.Requests.Storage().String())
	// 	}
	// }

	// for _, d := range data.Items {

	// var es esv1.Elasticsearch
	// es.Name = d.Object["metadata"].(map[string]interface{})["name"].(string)
	// es.Status = d.Object["status"].(map[string]interface{})["health"].(string)

	// nodeSetsJson := d.Object["spec"].(map[string]interface{})["nodeSets"].(string)

	// for _, nodeSet := range nodeSets.Items {
	// 	fmt.Printf("NodeSet Name : %s", nodeSet.Object["name"])
	// 	fmt.Printf("NodeSet Count : %s", nodeSet.Object["count"])
	// }
	// fmt.Printf("name: %s", es.Name)
	// fmt.Printf("status: %s", es.Status)
	// esList = append(esList, es)
	// }

	return true
}

func (controller *accountController) ParseToken(tokenString string) (string, bool) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		fmt.Println(err)
		return "", false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// var a := claims["kubernetes.io/serviceaccount/service-account.name"]
		return claims["kubernetes.io/serviceaccount/service-account.name"].(string), true
	} else {
		fmt.Println(err)
	}
	return "", false
}

func (controller *accountController) GetLoginAccount(c echo.Context) error {
	return c.JSON(http.StatusOK, controller.session.GetAccount())
}

func (controller *accountController) Logout(c echo.Context) error {
	sess := controller.GetSession()
	_ = sess.SetAccount(nil)
	_ = sess.Delete()
	return c.NoContent(http.StatusOK)
}
