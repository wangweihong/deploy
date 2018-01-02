package kv

import (
	"fmt"
	"strings"
	"ufleet-deploy/util/etcd3"

	"github.com/coreos/etcd/clientv3"
)

func NewEtcdV3Store(etcdHost string) (KVStore, error) {
	ec, err := etcd3.InitEtcd3Client(etcdHost)
	if err != nil {
		return nil, err
	}
	return &kvStoreV3{client: ec}, nil
}

type kvStoreV3 struct {
	client *etcd3.Etcd3Client
}

func (k *kvStoreV3) WatchNode(key string) (chan WatchEvent, error) {

	wechan := make(chan WatchEvent, 1000)
	watcherChan, err := k.client.GenerateWatcherChan(key)
	if err != nil {
		return nil, fmt.Errorf("generate etcd watcher fail for %v", err)
	}

	go func(wechan chan WatchEvent) {

		for wresp := range watcherChan {
			for _, ev := range wresp.Events {
				go func(ev *clientv3.Event) {
					var we WatchEvent
					switch ev.Type.String() {
					case "DELETE":
						we.Action = ActionDelete
					default:
						we.Action = ActionCreate
					}
					var node Node
					node.Key = string(ev.Kv.Key)
					node.Value = string(ev.Kv.Value)

					we.Node = &node
					wechan <- we
				}(ev)
			}
		}
	}(wechan)
	return wechan, nil
}

func (e *kvStoreV3) GetNode(key string) (*Node, error) {
	kvlist, err := e.client.Get(key)
	if err != nil {
		return nil, err
	}

	if len(kvlist) == 0 {
		return nil, ErrKeyNotFound
	}

	var node Node
	node.Key = kvlist[0].Key
	node.Value = string(kvlist[0].Value)

	return &node, nil
}

func (e *kvStoreV3) GetChildNode(key string) ([]Node, error) {
	kls, err := e.client.GetAll(key)
	if err != nil {
		return nil, err
	}

	if len(kls) == 0 {
		return nil, ErrKeyNotFound
	}

	nodes := make([]Node, 0)

	keymap := make(map[string]struct{})
	for _, v := range kls {
		if v.Key == key {
			continue
		}
		s := strings.TrimPrefix(v.Key, key+"/")
		t := strings.Split(s, "/")
		if _, ok := keymap[t[0]]; ok {
			continue
		}
		keymap[t[0]] = struct{}{}
		var n Node
		n.Key = v.Key
		n.Value = string(v.Value)
		nodes = append(nodes, n)
	}

	return nodes, nil
}

func (k *kvStoreV3) DeleteDirNode(key string) error {
	_, err := k.client.Delete(key+"/", true)
	if err != nil {
		return err
	}
	_, err = k.client.Delete(key, false)
	return err
}

func (k *kvStoreV3) DeleteNode(key string) error {
	fmt.Println("start ot delet enode ", key)
	_, err := k.client.Delete(key, false)
	return err
}

func (k *kvStoreV3) CreateDirNode(key string) error {
	_, err := k.client.Set(key, "")
	return err

}

func (k *kvStoreV3) CreateNode(key string, value interface{}) error {
	_, err := k.client.Set(key, value)
	return err
}

func (k *kvStoreV3) UpdateNode(key string, value interface{}) error {
	_, err := k.client.Set(key, value)
	return err
}

func (k *kvStoreV3) TestConnection() error {
	_, err := k.GetNode("/")
	if err != nil && err != ErrKeyNotFound {
		return err
	}

	return nil
}
