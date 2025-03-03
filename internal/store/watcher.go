package store

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func NewWatcher(options Options) (clientv3.Watcher, error) {
	if len(options.Endpoints) == 0 {
		options.Endpoints = DefaultOptions.Endpoints
	}

	config := clientv3.Config{
		Endpoints:   options.Endpoints,
		DialTimeout: 3 * time.Second,
	}

	cli, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	return clientv3.NewWatcher(cli), nil
}
