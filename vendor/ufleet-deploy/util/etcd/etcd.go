package etcd

import (
	"strings"
	"time"

	eclient "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

type EtcdClient struct {
	eclient.KeysAPI
}

func InitEtcdClient(endpoint string) *EtcdClient {
	if !strings.HasPrefix(endpoint, "http://") {
		endpoint = "http://" + endpoint
	}

	cfg := eclient.Config{
		Endpoints:               []string{endpoint},
		Transport:               eclient.DefaultTransport,
		HeaderTimeoutPerRequest: 10 * time.Second,
	}

	c, err := eclient.New(cfg)
	if err != nil {
		return nil
	}

	kapi := eclient.NewKeysAPI(c)
	etcdClient := &EtcdClient{
		kapi,
	}
	return etcdClient
}

func (e *EtcdClient) GetNode(key string) (*eclient.Response, error) {
	resp, err := e.Get(context.Background(), key, &eclient.GetOptions{Recursive: true})
	return resp, err
}

func (e *EtcdClient) GenerateWatcher(key string) (eclient.Watcher, error) {

	//FIXME
	//http://coreos.com/etcd/docs/latest/v2/api.html
	//http://hustcat.github.io/watch_in_etcd/
	watcher := e.Watcher(key, &eclient.WatcherOptions{Recursive: true, AfterIndex: 0})

	return watcher, nil

}

func (e *EtcdClient) DeleteDirNode(key string) (*eclient.Response, error) {
	resp, err := e.Delete(context.Background(), key, &eclient.DeleteOptions{Recursive: true})
	return resp, err
}

func (e *EtcdClient) DeleteNode(key string) (*eclient.Response, error) {
	resp, err := e.Delete(context.Background(), key, nil)
	return resp, err
}

func (e *EtcdClient) CreateDirNode(key, value string) (*eclient.Response, error) {
	resp, err := e.Set(context.Background(), key, value, &eclient.SetOptions{Dir: true, PrevExist: eclient.PrevNoExist})
	return resp, err
}

//存在则报错
func (e *EtcdClient) CreateNode(key, value string) (*eclient.Response, error) {
	resp, err := e.Set(context.Background(), key, value, &eclient.SetOptions{PrevExist: eclient.PrevNoExist})
	return resp, err
}

//不存在则报错
func (e *EtcdClient) UpdateNode(key, value string) (*eclient.Response, error) {
	resp, err := e.Set(context.Background(), key, value, &eclient.SetOptions{PrevExist: eclient.PrevExist})
	return resp, err
}
