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

var (
	Store    KVStore
	etcdHost string
)
var (
	ErrKeyNotFound      = fmt.Errorf("key not found")
	ErrKeyAlreadyExists = fmt.Errorf("key already exists")
)

type KVStore interface {
	GetNode(key string) (*Response, error)
	WatchNode(key string) (chan WatcheEvent, error)
	DeleteDirNode(key string) (*Response, error)
	DeleteNode(key string) (*Response, error)
	CreateDirNode(key string) (*Response, error)
	CreateNode(key string, value interface{}) (*Response, error)
	UpdateNode(key string, value interface{}) (*Response, error)
}

type Response struct {
	eclient.Response
}
type kvStore struct {
	client *etcd.EtcdClient
}

func NewKewStore(etcdHost string) KVStore {
	ec := etcd.InitEtcdClient(etcdHost)
	if ec == nil {
		return nil
	}
	return &kvStore{client: ec}

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

	Store = NewKewStore(etcdHost)
	if Store == nil {
		panic(fmt.Sprintf("init kvstore client for host \"%v\" fail", etcdHost))
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

func (k *kvStore) GetNode(key string) (*Response, error) {
	resp, err := k.client.GetNode(key)
	if err != nil {
		err = checkEtcdError(err)
		return nil, err
	}

	return &Response{*resp}, nil
}

type WatcheEvent struct {
	Resp *Response
	Err  error
}

func (k *kvStore) WatchNode(key string) (chan WatcheEvent, error) {
	wechan := make(chan WatcheEvent)
	watcher, err := k.client.GenerateWatcher(key)
	if err != nil {
		return nil, fmt.Errorf("generate etcd watcher fail for %v", err)
	}

	go func(wechan chan WatcheEvent) {
		count := 0
		for {

			resp, err := watcher.Next(context.Background())

			if err != nil {
				if count/20 == 0 {
					count = 0
					log.ErrorPrint("watch node fail for ", err)

				}
				count += 1
				time.Sleep(1 * time.Second)
				continue
			}
			wechan <- WatcheEvent{Err: err, Resp: &Response{*resp}}
		}
	}(wechan)
	return wechan, nil
}

func (k *kvStore) DeleteDirNode(key string) (*Response, error) {
	resp, err := k.client.DeleteDirNode(key)
	if err != nil {
		err = checkEtcdError(err)
		return nil, err
	}
	return &Response{*resp}, nil
}

func (k *kvStore) DeleteNode(key string) (*Response, error) {
	resp, err := k.client.DeleteNode(key)
	if err != nil {
		err = checkEtcdError(err)
		return nil, err
	}

	return &Response{*resp}, nil
}

func (k *kvStore) CreateDirNode(key string) (*Response, error) {
	resp, err := k.client.CreateDirNode(key, "")
	if err != nil {
		err = checkEtcdError(err)
		return nil, err
	}
	return &Response{*resp}, nil
}

func (k *kvStore) CreateNode(key string, value interface{}) (*Response, error) {
	byteContent, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	resp, err := k.client.CreateNode(key, string(byteContent))
	if err != nil {
		err = checkEtcdError(err)
		return nil, err
	}

	return &Response{*resp}, nil
}

func (k *kvStore) UpdateNode(key string, value interface{}) (*Response, error) {
	byteContent, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	resp, err := k.client.UpdateNode(key, string(byteContent))
	if err != nil {
		err = checkEtcdError(err)
		return nil, err
	}
	return &Response{*resp}, nil
}
