package controllers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"ufleet-deploy/models"
	"ufleet-deploy/pkg/resource"
	pk "ufleet-deploy/pkg/resource/secret"
	"ufleet-deploy/pkg/user"

	yaml "gopkg.in/yaml.v2"

	corev1 "k8s.io/client-go/pkg/api/v1"
)

type SecretController struct {
	baseController
}

// ListSecrets
// @Title Secret
// @Description  Secret
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Get]
func (this *SecretController) ListSecrets() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	pis, err := pk.Controller.ListGroupWorkspaceObject(group, workspace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetSecretInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}
	this.normalReturn(jss)
}

// ListGroupsSecrets
// @Title Secret
// @Description   Secret
// @Param Token header string true 'Token'
// @Param body body string true "组数组"
// @Success 201 {string} create success!
// @Failure 500
// @router /groups [Post]
func (this *SecretController) ListGroupsSecrets() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	groups := make([]string, 0)
	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit groups name")
		this.errReturn(err, 500)
		return
	}

	err := json.Unmarshal(this.Ctx.Input.RequestBody, &groups)
	if err != nil {
		err = fmt.Errorf("try to unmarshal data \"%v\" fail for %v", string(this.Ctx.Input.RequestBody), err)
		this.errReturn(err, 500)
		return
	}

	pis := make([]resource.Object, 0)

	for _, v := range groups {
		tmp, err := pk.Controller.ListGroupObject(v)
		if err != nil {
			this.errReturn(err, 500)
			return
		}
		pis = append(pis, tmp...)
	}
	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetSecretInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}
	this.normalReturn(jss)
}

// ListGroupSecrets
// @Title Secret
// @Description   Secret
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group [Get]
func (this *SecretController) ListGroupSecrets() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	pis, err := pk.Controller.ListGroupObject(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	jss := make([]pk.Status, 0)
	for _, j := range pis {
		v, _ := pk.GetSecretInterface(j)
		js := v.GetStatus()
		jss = append(jss, *js)
	}
	this.normalReturn(jss)
}

// CreateSecret
// @Title Secret
// @Description  创建私秘凭据
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace [Post]
func (this *SecretController) CreateSecret() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	var co models.CreateOption
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &co)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	ui := user.NewUserClient(token)
	who, err := ui.GetUserName()
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	var opt resource.CreateOption
	opt.Comment = co.Comment
	opt.User = who
	err = pk.Controller.CreateObject(group, workspace, []byte(co.Data), opt)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

type SecretCustomOption struct {
	Name           string `json:"name"`
	Comment        string `json:"comment,omitempty"`
	Type           string `json:"type"`
	Data           string `json:"data,omitempty"`
	Registry       string `json:"registry,omitempty"`
	ServiceAccount string `json:"serviceaccount"`
}

type DockerRegistryAccount struct {
	User     string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email'`
	Auth     string `json:"auth"`
}

// CreateSecretCustom
// @Title Secret
// @Description  创建私秘凭据
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /group/:group/workspace/:workspace/custom [Post]
func (this *SecretController) CreateSecretCustom() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	ui := user.NewUserClient(token)
	who, err := ui.GetUserName()
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	var co SecretCustomOption
	err = json.Unmarshal(this.Ctx.Input.RequestBody, &co)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	cm := corev1.Secret{}
	cm.Name = co.Name
	cm.Kind = "Secret"
	cm.APIVersion = "v1"
	switch corev1.SecretType(co.Type) {
	case corev1.SecretTypeDockercfg:
		cm.Type = corev1.SecretTypeDockercfg

		if co.Data == "" {
			reg, err := ui.GetRegistry(group, co.Registry)
			if err != nil {
				this.audit(token, "", true)
				this.errReturn(err, 500)
				return
			}

			account := make(map[string]DockerRegistryAccount)
			ra := DockerRegistryAccount{
				User:     reg.User,
				Password: reg.Password,
				Email:    reg.Email,
			}

			originAuthString := reg.User + ":" + reg.Password
			ra.Auth = base64.StdEncoding.EncodeToString([]byte(originAuthString))

			account[reg.Address] = ra

			dockercfg, err := json.Marshal(account)
			if err != nil {
				this.audit(token, "", true)
				this.errReturn(err, 500)
				return

			}

			cm.Data = make(map[string][]byte)
			cm.Data[corev1.DockerConfigKey] = []byte(dockercfg)

		} else {
			/*
					调用示例:

					{
					"name":"haha2",
					"type":"kubernetes.io/dockercfg",
					"data": ".dockercfg: eyIxOTIuMTY4LjE0LjEwMDo1MDAyIjp7InVzZXJuYW1lIjoiYWRtaW4iLCJwYXNzd29yZCI6IjEyMzQ1NiIsImVtYWlsIjoiZW1hbEAxMjMuY29tIiwiYXV0aCI6IllXUnRhVzQ2TVRJek5EVTIifX0="
				}
			*/

			//直接通过map[string][]byte无法进行解析.在这里就会报错
			data := make(map[string]string)
			err := yaml.Unmarshal([]byte(co.Data), &data)
			if err != nil {
				this.audit(token, "", true)
				this.errReturn(err, 500)
				return
			}

			cm.Data = make(map[string][]byte)
			//不能直接通过cm.Data[k]= []byte(v),在创建时会报
			//Error:Secret "werr" is invalid: data[.dockercfg]: Invalid value: "<secret contents redacted>": invalid character 'e' looking for beginning of value
			//即使里面的值为打印后为"eyIxOTIuMTY4LjE0LjEwMDo1MDAyIjp7InVzZXJuYW1lIjoiYWRtaW4iLCJwYXNzd29yZCI6IjEyMzQ1NiIsImVtYWlsIjoiZW1hbEAxMjMuY29tIiwiYXV0aCI6IllXUnRhVzQ2TVRJek5EVTIifX0="
			for k, v := range data {
				//base64加密过需要解码

				d, err := base64.StdEncoding.DecodeString(v)
				if err != nil {
					this.audit(token, "", true)
					this.errReturn(err, 500)
					return
				}

				cm.Data[k] = d
			}

		}

	case corev1.SecretTypeServiceAccountToken:
		if strings.TrimSpace(co.ServiceAccount) == "" {
			err := fmt.Errorf("secret type '%v' must offer service account name", co.Type)
			this.audit(token, "", true)
			this.errReturn(err, 500)
			return

		}
		cm.Annotations = make(map[string]string)
		cm.Annotations[corev1.ServiceAccountNameKey] = co.ServiceAccount
		cm.Type = corev1.SecretType(co.Type)

	default:
		if strings.TrimSpace(co.Data) == "" {
			err := fmt.Errorf("must offer secret data")
			this.audit(token, "", true)
			this.errReturn(err, 500)
			return
		}

		data := make(map[string]string)
		err := yaml.Unmarshal([]byte(co.Data), &data)
		if err != nil {
			this.audit(token, "", true)
			this.errReturn(err, 500)
			return
		}

		cm.Type = corev1.SecretType(co.Type)
		cm.StringData = data
	}

	bytedata, err := json.Marshal(cm)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	var opt resource.CreateOption
	opt.Comment = co.Comment
	opt.User = who
	err = pk.Controller.CreateObject(group, workspace, bytedata, opt)
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// UpdateSecret
// @Title Secret
// @Description  更新secret
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param secret path string true "secret"
// @Param body body string true "资源描述"
// @Success 201 {string} create success!
// @Failure 500
// @router /:secret/group/:group/workspace/:workspace [Put]
func (this *SecretController) UpdateSecret() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	//token := this.Ctx.Request.Header.Get("token")
	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	secret := this.Ctx.Input.Param(":secret")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit resource json/yaml data")
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}

	err := pk.Controller.UpdateObject(group, workspace, secret, this.Ctx.Input.RequestBody, resource.UpdateOption{})
	if err != nil {
		this.errReturn(err, 500)
		this.audit(token, "", true)
		return
	}

	this.audit(token, "", false)
	this.normalReturn("ok")
}

// DeleteSecret
// @Title Secret
// @Description   Secret
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param secret path string true "私秘凭据"
// @Success 201 {string} create success!
// @Failure 500
// @router /:secret/group/:group/workspace/:workspace [Delete]
func (this *SecretController) DeleteSecret() {
	token := this.Ctx.Request.Header.Get("token")
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.audit(token, "", true)
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	secret := this.Ctx.Input.Param(":secret")

	err := pk.Controller.DeleteObject(group, workspace, secret, resource.DeleteOption{})
	if err != nil {
		this.audit(token, "", true)
		this.errReturn(err, 500)
		return
	}
	this.audit(token, "", false)

	this.normalReturn("ok")
}

// GetSecretTemplate
// @Title Secret
// @Description   Secret
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param secret path string true "私秘凭据"
// @Success 201 {string} create success!
// @Failure 500
// @router /:secret/group/:group/workspace/:workspace/template [Get]
func (this *SecretController) GetSecretTemplate() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	secret := this.Ctx.Input.Param(":secret")

	v, err := pk.Controller.GetObject(group, workspace, secret)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetSecretInterface(v)

	t, err := pi.GetTemplate()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(t)
}

// GetSecretContainerEvents
// @Title Secret
// @Description   Secret container event
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param workspace path string true "工作区"
// @Param secret path string true "私秘凭据"
// @Success 201 {string} create success!
// @Failure 500
// @router /:secret/group/:group/workspace/:workspace/event [Get]
func (this *SecretController) GetSecretEvent() {
	aerr := this.checkRouteControllerAbility()
	if aerr != nil {
		this.abilityErrorReturn(aerr)
		return
	}

	group := this.Ctx.Input.Param(":group")
	workspace := this.Ctx.Input.Param(":workspace")
	endpoint := this.Ctx.Input.Param(":endpoint")

	v, err := pk.Controller.GetObject(group, workspace, endpoint)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	pi, _ := pk.GetSecretInterface(v)
	es, err := pi.Event()
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(es)
}
