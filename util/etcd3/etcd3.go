package etcd3

import (
	"context"
	"strings"
	"time"

	"encoding/json"

	"github.com/coreos/etcd/clientv3"
)

/*
var (
	etcdClient Etcd3Client
)
*/

func InitEtcd3Client(etcdCluster string) (*Etcd3Client, error) {

	endpoints := strings.Split(etcdCluster, ",")
	etcdClient := Etcd3Client{
		endpoints,
		5,
		nil,
	}
	err := etcdClient.newClient()
	if err != nil {
		return nil, err
	}
	return &etcdClient, nil

}

// Etcd3Client  etcd客户端
type Etcd3Client struct {
	endpoints   []string
	dialTimeout int
	RawClient   *clientv3.Client
}

// newClient 获取key
func (e *Etcd3Client) newClient() error {
	requestTimeout := 5 * time.Second

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   e.endpoints,
		DialTimeout: requestTimeout,
	})
	if err != nil {
		return err
	}
	e.RawClient = cli
	return nil
}

// KeyList etcd
type KeyList struct {
	Key   string
	Value []byte
}

// Get 获取key下的所有目录
func (e *Etcd3Client) Get(key string, opt ...clientv3.OpOption) ([]KeyList, error) {
	var value []KeyList
	response, err := e.RawClient.Get(context.Background(), key, opt...)
	if err != nil {
		return value, err
	}

	for _, item := range response.Kvs {
		itemKey := string(item.Key)
		value = append(value, KeyList{itemKey, item.Value})
	}

	return value, err
}

func (e *Etcd3Client) GetAll(key string) ([]KeyList, error) {
	values, err := e.Get(key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	return values, err
}

func (e *Etcd3Client) GetKeyChild(key string) ([]KeyList, error) {

	kls, err := e.GetAll(key)
	if err != nil {
		return nil, err
	}

	ckls := make([]KeyList, 0)
	for _, v := range kls {
		s := strings.TrimPrefix(v.Key, key)
		t := strings.Split(s, "/")
		if len(t) == 0 || s != "" {
			ckls = append(ckls, v)
		}
	}
	return ckls, nil
}

// Set 保存key到etcd中,可以接受struct,数组,map,string
func (e *Etcd3Client) Set(key string, value interface{}, opt ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	var response *clientv3.PutResponse
	var err error
	var data string

	switch t := value.(type) {
	case string:
		data = t
	case interface{}:
		{
			b, err := json.Marshal(value)
			if err != nil {
				return response, err
			}
			data = string(b)
		}
	case byte:
		data = string(t)
	}

	response, err = e.RawClient.Put(context.Background(), key, data, opt...)
	return response, err
}

/*
func (e *Etcd3Client) Get(key string, opt ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	response, err = e.RawClient.Get(context.Background(), key, opt...)
	return response, nil
}
*/

// Delete 删除key
func (e *Etcd3Client) Delete(key string, recursive bool, opt ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	var response *clientv3.DeleteResponse
	var err error

	if recursive {
		response, err = e.RawClient.Delete(context.Background(), key, clientv3.WithPrefix())
	} else {
		response, err = e.RawClient.Delete(context.Background(), key, opt...)
	}
	return response, err
}

func (e *Etcd3Client) GenerateWatcherChan(key string) (clientv3.WatchChan, error) {

	watcher := e.RawClient.Watch(context.Background(), key, clientv3.WithPrefix())

	return watcher, nil

}
