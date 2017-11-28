package cluster

import (
	"encoding/json"
	"fmt"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/util/request"

	rest "k8s.io/client-go/rest"
)

var (
	clusterHost string
	hostDomain  string
)

type ClusterConfig struct {
	Auth_way  string `json:"auth_way"`
	Auth_url  string `json:"auth_url"`
	Workspace string `json:"workspace"`
	//basic authentication
	User     string `json:"user"`
	Password string `json:"password"`
	//
	ApiServer string `json:"server"`
	//TLS authentication
	TLSCAData   string `json:"auth_data"`
	TLSCertData string `json:"cert_data"`
	TLSKeyData  string `json:"client_key"`

	//bearer authentication
	BearerToken string `json:"auth_url"`
	ClusterName string `json:"cluster_name"`
}

type Response struct {
	Result  int             `json:"result"`
	Message string          `json:"message"`
	Content json.RawMessage `json:"content"`
}

func GetK8sClientConfig(group string, workspace string, token string) (*ClusterConfig, error) {

	var url string
	url = clusterHost + "/v1/cluster/authenticat/?group_name=" + group + "&workspace=" + workspace
	//url
	body, err := request.Get(url, token)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	var resp Response
	err = json.Unmarshal(body, &resp)
	if err != nil {
		log.ErrorPrint("get clutster info :", string(body))
		return nil, log.ErrorPrint(err)
	}
	if resp.Result != 0 {
		return nil, log.ErrorPrint(fmt.Errorf("%v", resp))
	}

	//	log.ErrorPrint(string(resp.Content))

	var wc WorkspaceAndConfig
	err = json.Unmarshal(resp.Content, &wc)
	if err != nil {
		err = log.ErrorPrint(fmt.Errorf("unmarshal [%v] (group '%v', workspace '%v') to client config fail for %v", string(resp.Content), group, workspace, err))
		return nil, err
	}

	if len(wc.Config.ApiServer) == 0 || len(wc.Config.Auth_way) == 0 {
		return nil, log.ErrorPrint(fmt.Errorf("can't get cluster info for invalid apiserver [%v] or auth way [%v]", wc.Config.ApiServer, wc.Config.Auth_way))
	}

	//	return clientConfigToK8s(wc.Config)
	return &wc.Config, nil

}

func ClusterConfigToK8sClientConfig(c ClusterConfig) (*rest.Config, error) {
	//uerr.PrintAndReturnError(c.ClusterName)
	var rconfig rest.Config
	rconfig.Host = c.ApiServer

	switch c.Auth_way {
	case "ca_auth":
		var tlsConfig rest.TLSClientConfig
		tlsConfig.CAData = []byte(c.TLSCAData)
		tlsConfig.KeyData = []byte(c.TLSKeyData)
		tlsConfig.CertData = []byte(c.TLSCertData)
		rconfig.TLSClientConfig = tlsConfig
	case "basic":
		rconfig.Host = c.ApiServer
		rconfig.Username = c.User
		rconfig.Password = c.Password
	case "auth_url":
		rconfig.BearerToken = c.BearerToken
		//8080无须认证
	case "http":
		rconfig.Host = c.ApiServer

	default:
		return nil, log.DebugPrint(fmt.Errorf("auth config %v can't get active authentication way:%v", c, c.Auth_way))

	}
	return &rconfig, nil
}

type WorkspaceAndConfig struct {
	Config    ClusterConfig `json:"k8sconf"`
	Namespace string        `json:"namespace"`
}

func GetTerminalUrl(group, workspace string, podName string, containerName string, hostIp string, clusterName string, token string) (string, error) {
	url := hostDomain + "/console/?cluster_name=" + clusterName + "&" + "pod_name=" + podName + "&" + "container_name=" + containerName + "&" + "namespace=" + workspace + "&" + "ip=" + hostIp
	return url, nil
}
