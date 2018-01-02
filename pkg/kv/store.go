package kv

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"ufleet-deploy/pkg/log"
	"ufleet-deploy/util/etcd"

	eclient "github.com/coreos/etcd/client"
)

const (
	ActionDelete = "delete"
	ActionCreate = "set"
)

var (
	Store    KVStore
	etcdHost string

	ErrKeyNotFound      = fmt.Errorf("key not found")
	ErrKeyAlreadyExists = fmt.Errorf("key already exists")
)

type KVStore interface {
	GetNode(key string) (*Node, error)
	GetChildNode(key string) ([]Node, error)
	WatchNode(key string) (chan WatchEvent, error)
	DeleteDirNode(key string) error
	DeleteNode(key string) error
	CreateDirNode(key string) error
	CreateNode(key string, value interface{}) error
	UpdateNode(key string, value interface{}) error
	TestConnection() error
}

type Node struct {
	//	Action string `json:"action"`
	Key   string `json:"key"`
	Value string `json:"value"`
	TTL   int64  `json:"ttl"`
}

type Response struct {
	//eclient.Response
	Action string `json:"action"`
	Node   *Node
}
type kvStore struct {
	client *etcd.EtcdClient
}

func NewKewStore(etcdHost string) (KVStore, error) {
	//	return NewEtcdStore(etcdHost)
	return NewEtcdV3Store(etcdHost)
}

func NewEtcdStore(etcdHost string) (KVStore, error) {
	ec, err := etcd.InitEtcdClient(etcdHost)
	if err != nil {
		return nil, err
	}
	return &kvStore{client: ec}, nil

}

func GetKVStoreAddr() string {
	return etcdHost
}

func Init(etcdHostEnvKey string) {
	etcdHost = etcdHostEnvKey
	//etcd删除
	/*
		etcdHost = os.Getenv(etcdHostEnvKey)
		if len(etcdHost) == 0 {
			panic(fmt.Sprintf("must provide Environment \"%v\"", etcdHostEnvKey))
		}
	*/

	var err error
	Store, err = NewKewStore(etcdHost)
	if err != nil {
		panic(fmt.Sprintf("init kvstore client for host \"%v\" fail for '%v'", etcdHost, err))
	}

}

func checkEtcdError(err error) error {
	if e, ok := err.(eclient.Error); ok {
		if e.Code == eclient.ErrorCodeKeyNotFound {
			return ErrKeyNotFound
		}
		if e.Code == eclient.ErrorCodeNodeExist {
			return ErrKeyAlreadyExists
		}
	}
	return err
}

func (k *kvStore) GetNode(key string) (*Node, error) {
	eresp, err := k.client.GetNode(key)
	if err != nil {
		err = checkEtcdError(err)
		return nil, err
	}

	//	resp := etcdRespondeToKvResponse(*eresp)
	var node Node
	node.Key = eresp.Node.Key
	node.Value = eresp.Node.Value
	node.TTL = eresp.Node.TTL

	return &node, nil
}

func (k *kvStore) GetChildNode(key string) ([]Node, error) {
	eresp, err := k.client.GetNode(key)
	if err != nil {
		err = checkEtcdError(err)
		return nil, err
	}

	nodes := make([]Node, 0)
	for _, v := range eresp.Node.Nodes {
		var node Node
		node.Key = v.Key
		node.Value = v.Value
		node.TTL = v.TTL
		nodes = append(nodes, node)
	}

	return nodes, nil
}

type WatchEvent struct {
	//	Resp *Response
	Node   *Node
	Action string
	Err    error
}

func (k *kvStore) WatchNode(key string) (chan WatchEvent, error) {
	wechan := make(chan WatchEvent)
	watcher, err := k.client.GenerateWatcher(key)
	if err != nil {
		return nil, fmt.Errorf("generate etcd watcher fail for %v", err)
	}

	go func(wechan chan WatchEvent) {
		count := 0
		for {

			eresp, err := watcher.Next(context.Background())

			if err != nil {
				if count/20 == 0 {
					count = 0
					log.ErrorPrint("watch node fail for ", err)

				}
				count += 1
				time.Sleep(1 * time.Second)
				continue
			}

			//			resp := etcdRespondeToKvResponse(*eresp)
			var node Node
			node.Key = eresp.Node.Key
			node.Value = string(eresp.Node.Value)

			wechan <- WatchEvent{Err: err, Node: &node, Action: eresp.Action}
		}
	}(wechan)
	return wechan, nil
}

//func (k *kvStore) DeleteDirNode(key string) (*Response, error) {
func (k *kvStore) DeleteDirNode(key string) error {
	_, err := k.client.DeleteDirNode(key)
	if err != nil {
		err = checkEtcdError(err)
		return err
	}
	return nil
}

//func (k *kvStore) DeleteNode(key string) (*Response, error) {
func (k *kvStore) DeleteNode(key string) error {
	_, err := k.client.DeleteNode(key)
	if err != nil {
		err = checkEtcdError(err)
		return err
	}

	return nil
}

//func (k *kvStore) CreateDirNode(key string) (*Response, error) {
func (k *kvStore) CreateDirNode(key string) error {
	_, err := k.client.CreateDirNode(key, "")
	if err != nil {
		err = checkEtcdError(err)
		return err
	}
	return nil
}

//func (k *kvStore) CreateNode(key string, value interface{}) (*Response, error) {
func (k *kvStore) CreateNode(key string, value interface{}) error {
	byteContent, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = k.client.CreateNode(key, string(byteContent))
	if err != nil {
		err = checkEtcdError(err)
		return err
	}

	return nil
}

//func (k *kvStore) UpdateNode(key string, value interface{}) (*Response, error) {
func (k *kvStore) UpdateNode(key string, value interface{}) error {
	byteContent, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = k.client.UpdateNode(key, string(byteContent))
	if err != nil {
		err = checkEtcdError(err)
		return err
	}
	return nil
}

func (k *kvStore) TestConnection() error {
	return k.client.TestConnection()
}
