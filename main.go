/*
Copyright 2016 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nikawang/es-gen-backend/controller"
	"github.com/nikawang/es-gen-backend/sessionss"
)

// "flag"

// "path/filepath"

// "k8s.io/client-go/tools/clientcmd"
// "k8s.io/client-go/util/homedir"
// Uncomment to load all auth plugins
// _ "k8s.io/client-go/plugin/pkg/client/auth"
//
// Or uncomment to load specific auth plugins
// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

func SessionInit(sess sessionss.Session) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sess.SetContext(c)
			if err := next(c); err != nil {
				c.Error(err)
			}
			return nil
		}
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowCredentials: true,
		AllowOrigins:     []string{"*"},
		AllowHeaders: []string{
			echo.HeaderAccessControlAllowHeaders,
			echo.HeaderContentType,
			echo.HeaderContentLength,
			echo.HeaderAcceptEncoding,
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
		},
		MaxAge: 86400,
	}))

	sess := sessionss.NewSession()

	accountController := controller.NewAccountController(sess)
	e.POST(controller.APIAccountLogin, func(c echo.Context) error { return accountController.Login(c) })
	e.GET(controller.APIAccountLoginStatus, func(c echo.Context) error { return accountController.GetLoginStatus(c) })
	e.GET(controller.APIAccountLoginAccount, func(c echo.Context) error { return accountController.GetLoginAccount(c) })
	e.POST(controller.APIAccountLogout, func(c echo.Context) error { return accountController.Logout(c) })

	eSController := controller.NewESController(sess)
	e.GET(controller.APIES, func(c echo.Context) error { return eSController.GetESList(c) })
	e.GET(controller.APIESNAME, func(c echo.Context) error { return eSController.GetES(c) })
	e.PUT(controller.APIES, func(c echo.Context) error { return eSController.UpdateES(c) })
	e.POST(controller.APIES, func(c echo.Context) error { return eSController.CreateES(c) })
	e.DELETE(controller.APIESNAME, func(c echo.Context) error { return eSController.DeleteES(c) })

	e.Use(SessionInit(sess))
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Logger.Fatal(e.Start(":8080"))

	// t.Printf("testst1")
	// token := "eyJhbGciOiJSUzI1NiIsImtpZCI6IllxRkhiLTg5QTVJNnRDQTg3ZjI0SDVsOFFpSlNuWE5ubXl4emtXblRHZW8ifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJrdWJlLXN5c3RlbSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VjcmV0Lm5hbWUiOiJpbnRlcm5hbGFkbWluLXRva2VuLXpid3RxIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6ImludGVybmFsYWRtaW4iLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiJmYTQ1OTQ3Yi1lNzg3LTQyNDQtODlkYS00Y2FjZTQ1Mzg0MzEiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6a3ViZS1zeXN0ZW06aW50ZXJuYWxhZG1pbiJ9.D9FjDsxLS40lX70TerXaYkixOQiuntncFinwEYYRsl8Sx-kfAThjoyG4JWCYtgM7OKuNFO99_uQGedAHe8N07TVO27VCXAhQNJR3YL1m7PgX47wwr6sl2HFY5NDPDG_pEVbAGf2_9Jvj6TJCRaPXOmqECGaW8AbM1Sbkx65zcIhIhK4c7O9djbHvFdwUw7uagASj9jA7DaNXXjuAEtjCSbP_bTougepMZ4vlQ79U1T43iKRfyi45YqkSrltsDfRa3YqxciEkI6GUMTHqukRwFMw_Zr0G0LFR63pUg-n2o0fUnha1Gx3YVeBiybRaVy9dqkIuzHBl60tCwAho-FlPChyZs96YiI7baapPRKNAtxDmg0mtATCDbXOKA4tsozXmCoTEbOCR31KfnLCOYeUQ6uT6jZxLbI8WQkDI4TzToyfonc7ZY-ZskmVeINLfFHnQlowbxo8yDu6WUTT0a5DrXOb9F1-dCA0O_JlpdCmwdpVezrVhh34PCcCnmzn5rCmfc89VuXwHqBi-2YbkCuLSOkeAlyWVXKWHvFAYFfBfLM_G8T4m9rqsOhrpuQmrReuFE03HJynhthKvXec0HgkdPptsizjP-kztl0REvltYgCKBQQVj8KSeZ9-f1OxxHseKWeHDxGiYyd6YAXVUKvxC1UzTYjxyjz2WzX71FHrTdkQ"
	// host := "https://egressgw-aks1-77bbb8-df9525d7.hcp.westcentralus.azmk8s.io"
	// tlscfg := &rest.TLSClientConfig{
	// 	Insecure: true,
	// }
	// cfg := &rest.Config{
	// 	Host:            host,
	// 	BearerToken:     token,
	// 	TLSClientCo xnfig: *tlscfg,
	// }
	// clientset, err := kubernetes.NewForConfig(cfg)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// var kubeconfig *string
	// if home := homedir.HomeDir(); home != "" {
	// 	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	// } else {
	// 	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	// }
	// flag.Parse()

	// // use the current context in kubeconfig
	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	panic(err.Error())
	// }

	// clientset, err := kubernetes.NewForConfig(config)
	// pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	// if err != nil {
	// 	panic(err.Error())
	// }
	// fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	// fmt.Printf("testst")
	// pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	// if err != nil {
	// 	panic(err.Error())
	// }
	// fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

}
