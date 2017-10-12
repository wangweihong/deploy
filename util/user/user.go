package user

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"ufleet-deploy/util/request"
)

type Registry struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	User     string `json:"user"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

type UserInterface interface {
	GetRegistrysFromGroup(string, string) ([]Registry, error)
	GetUserName(string) (string, error)
}

type userClient struct {
	token string
}

func NewUserClient(token string) UserInterface {
	var uc userClient
	uc.token = token
	return &uc

}

func (uc *userClient) GetRegistrysFromGroup(UserIp string, groupName string) ([]Registry, error) {
	url := UserIp + "/v1/registry/group" + "/" + groupName
	data, err := request.Get(url, uc.token)
	if err != nil {
		return nil, err
	}
	//	fmt.Println(string(data))

	reg := make([]Registry, 0)
	err = json.Unmarshal(data, &reg)
	if err != nil {
		return nil, err
	}

	//需要base64解密
	for k, _ := range reg {
		rawPwd, err := base64.StdEncoding.DecodeString(reg[k].Password)
		if err != nil {
			return nil, err
		}

		reg[k].Password = string(rawPwd)
	}

	//	fmt.Println("=====================")
	//	fmt.Println(reg)
	/*
		var r Registry
		r.Name = "haha"
		r.Address = "192.168.18.250:5002"
		r.User = "admin"
		r.Password = "123456"
		reg = append(reg, r)
	*/

	return reg, nil

}

type User struct {
	Username string `json:"username"`
}

func (uc *userClient) GetUserName(UserIp string) (string, error) {

	url := UserIp + "/v1/user/verify/" + strings.TrimRight(uc.token, "/")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Token", uc.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("try to get user info fail for %v", string(body))
	}

	var u User
	err = json.Unmarshal(body, &u)
	if err != nil {
		return "", err
	}

	return u.Username, nil
}
