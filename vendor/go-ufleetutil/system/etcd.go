package system

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"encoding/json"

	"github.com/coreos/etcd/clientv3"
)

var (
	etcdClient Etcd3Client
)

func init() {
	etcdCluster := os.Getenv("ETCDCLUSTER")
	endpoints := strings.Split(etcdCluster, ",")
	etcdClient = Etcd3Client{
		endpoints,
		5,
		nil,
	}
	err := etcdClient.newClient()
	if err != nil {
		log.Fatalln(err)
	}
}

// NewEtcd3Client 实例化客户端,需要将etcd的地址配置成环境变量`ETCDCLUSTER`
func NewEtcd3Client() Etcd3Client {
	return etcdClient
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
