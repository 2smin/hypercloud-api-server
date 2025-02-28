package caller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tmax-cloud/hypercloud-api-server/util"

	"k8s.io/klog"
)

func setHyperAuthURL(serviceName string) string {
	hyperauthHttpPort := "8080"
	if os.Getenv("HYPERAUTH_HTTP_PORT") != "" {
		hyperauthHttpPort = os.Getenv("HYPERAUTH_HTTP_PORT")
	}
	return util.HYPERAUTH_URL + ":" + hyperauthHttpPort + "/" + serviceName
}

func LoginAsAdmin() string {
	klog.V(3).Infoln(" [HyperAuth] Login as Admin Service")
	// Make Body for Content-Type (application/x-www-form-urlencoded)
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", "admin")
	data.Set("password", "admin")
	data.Set("client_id", "admin-cli")

	// Make Request Object
	req, err := http.NewRequest("POST", setHyperAuthURL(util.HYPERAUTH_SERVICE_NAME_LOGIN_AS_ADMIN), strings.NewReader(data.Encode()))
	if err != nil {
		klog.V(1).Info(err)
		panic(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Request with Client Object
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		klog.V(1).Info(err)
		panic(err)
	}
	defer resp.Body.Close()

	// Result
	bytes, _ := ioutil.ReadAll(resp.Body)
	str := string(bytes) // byte to string
	klog.V(3).Infoln("Result string  : ", str)

	var resultJson map[string]interface{}
	if err := json.Unmarshal([]byte(str), &resultJson); err != nil {
		klog.V(1).Info(err)
	}
	accessToken := resultJson["access_token"].(string)
	return accessToken
}

// defunct
// func getUserDetailWithoutToken(userId string) map[string]interface{} {
// 	klog.V(3).Infoln(" [HyperAuth] HyperAuth Get User Detail Without Token Service")

// 	// Make Request Object
// 	req, err := http.NewRequest("GET", setHyperAuthURL(util.HYPERAUTH_SERVICE_NAME_USER_DETAIL_WITHOUT_TOKEN)+userId, nil)
// 	if err != nil {
// 		klog.V(1).Info(err)
// 		panic(err)
// 	}

// 	// Request with Client Object
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		klog.V(1).Info(err)
// 		panic(err)
// 	}
// 	defer resp.Body.Close()

// 	// Result
// 	bytes, _ := ioutil.ReadAll(resp.Body)
// 	str := string(bytes) // byte to string
// 	klog.V(3).Infoln("Result string  : ", str)

// 	var resultJson map[string]interface{}
// 	if err := json.Unmarshal([]byte(str), &resultJson); err != nil {
// 		klog.V(1).Info(err)
// 		panic(err)
// 	}
// 	return resultJson
// }
