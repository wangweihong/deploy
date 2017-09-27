package backend

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
	"ufleet-deploy/deploy/kv"

	"github.com/astaxie/beego"
)

//{"action":"get","node":{"key":"/ufleet/registry/g2","dir":true,"nodes":[{"key":"/ufleet/registry/g2/Reg20170806230821jP8C","value":"{\"id\":\"Reg20170806230821jP8C\",\"name\":\"reg\",\"address\":\"http://192.168.18.250:5002\",\"user\":\"admin\",\"password\":\"MTIzNDU2\",\"updateTime\":1502075301,\"belong\":\"g2\"}","modifiedIndex":19120,"createdIndex":19120}],"modifiedIndex":18757,"createdIndex":18757}}
const (
	etcdRegistryKey = "/ufleet/registry"
)

var (
	registryNoticers = make(map[string]chan RegistryEvent)
	regLock          = sync.Mutex{}
)

type Registry struct {
	ID       string `id`
	Name     string `name`
	Address  string `address`
	User     string `user`
	Password string `password`
	//	UpdateTime int64  `updateTime`
	Group string `belong`
}

type RegistryEvent struct {
	Action   string
	Group    string
	Registry *Registry
}

func RegisterRegistryNoticer(name string) (chan RegistryEvent, error) {
	regLock.Lock()
	defer regLock.Unlock()

	regChan := make(chan RegistryEvent)
	if _, ok := registryNoticers[name]; ok {
		return nil, fmt.Errorf("noticer \"%v\" has registered", name)
	}

	registryNoticers[name] = regChan
	return regChan, nil

}

//TODO:还需要监控registry更新事件
func watchRegistyEvent() error {
	wechan, err := kv.Store.WatchNode(etcdRegistryKey)
	if err != nil {
		return err
	}

	go func() {
		for {
			we := <-wechan
			if we.Err != nil {
				beego.Error(we.Err)
				time.Sleep(1 * time.Second)
				continue
			}

			res := we.Resp
			if res.Node.Key == etcdRegistryKey {
				continue
			}

			remain := strings.TrimPrefix(res.Node.Key, etcdRegistryKey+"/")
			s := strings.Split(remain, "/")
			var r *Registry
			//registry
			group := s[0]
			if len(s) == 2 {
				value := res.Node.Value
				err := json.Unmarshal([]byte(value), r)
				if err != nil {
					beego.Error(err)
					continue
				}
			}

			var event RegistryEvent
			event.Action = res.Action
			event.Group = group
			event.Registry = r

			for _, v := range registryNoticers {
				go func(c chan RegistryEvent) {
					v <- event
				}(v)
			}

		}
	}()

	return nil
}
