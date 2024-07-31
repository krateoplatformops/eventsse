package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	corev1 "k8s.io/api/core/v1"
)

type TTLSetter interface {
	SetTTL(ttl int)
}

type KeyPreparer interface {
	PrepareKey(eventId, compositionId string) string
}

type Closer interface {
	Close() error
}

var (
	defaultTimeout             = 200 * time.Millisecond
	_              TTLSetter   = (*Client)(nil)
	_              KeyPreparer = (*Client)(nil)
	_              Store       = (*Client)(nil)
)

type Store interface {
	TTLSetter
	KeyPreparer
	Closer
	Set(k string, v *corev1.Event) error
	Get(k string, opts GetOptions) (data []corev1.Event, found bool, err error)
	Delete(k string) error
}

// Client is a Store implementation for etcd.
type Client struct {
	c       *clientv3.Client
	timeOut time.Duration
	ttl     int
}

func (c *Client) SetTTL(ttl int) {
	c.ttl = ttl
}

func (c *Client) PrepareKey(eventId, compositionId string) string {
	key := ""
	if len(compositionId) > 0 {
		key = path.Join(key, fmt.Sprintf("comp-%s", compositionId))
	}
	if len(eventId) > 0 {
		key = path.Join(key, eventId)
	}
	key = path.Join("events", strings.ToLower(key))
	return key
}

// Set stores the given value for the given key.
func (c *Client) Set(k string, v *corev1.Event) error {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return err
	}

	opts := []clientv3.OpOption{}
	if c.ttl > 0 {
		lease := clientv3.NewLease(c.c)
		res, err := lease.Grant(context.Background(), int64(c.ttl))
		if err != nil {
			return err
		}
		opts = append(opts, clientv3.WithLease(res.ID))
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()
	_, err := c.c.Put(ctxWithTimeout, k, buf.String(), opts...)
	return err
}

type GetOptions struct {
	Limit  int
	EndKey string
}

// Get retrieves the stored value for the given key.
func (c *Client) Get(k string, opts GetOptions) (data []corev1.Event, found bool, err error) {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()

	ops := []clientv3.OpOption{
		clientv3.WithLimit(int64(opts.Limit)),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend),
	}
	if len(opts.EndKey) > 0 {
		ops = append(ops, clientv3.WithRange(opts.EndKey))
	} else {
		ops = append(ops, clientv3.WithPrefix())
	}

	getRes, err := c.c.Get(ctxWithTimeout, k, ops...)
	if err != nil {
		return data, false, err
	}

	kvs := getRes.Kvs
	// If no value was found return false
	if len(kvs) == 0 {
		return data, false, nil
	}

	for _, el := range kvs {
		var obj corev1.Event
		if err := json.Unmarshal(el.Value, &obj); err != nil {
			return data, false, err
		}

		data = append(data, obj)
	}

	return data, true, nil
}

// Delete deletes the stored value for the given key.
func (c *Client) Delete(k string) error {
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), c.timeOut)
	defer cancel()
	_, err := c.c.Delete(ctxWithTimeout, k)
	return err
}

// Close closes the client.
func (c *Client) Close() error {
	return c.c.Close()
}

// Options are the options for the etcd client.
type Options struct {
	// Addresses of the etcd servers in the cluster, including port.
	// Optional ([]string{"localhost:2379"} by default).
	Endpoints []string
	// Sored Items TTL in seconds
	TTL int64
}

// DefaultOptions is an Options object with default values.
// Endpoints: []string{"localhost:2379"}, Timeout: 200 * time.Millisecond, Codec: encoding.JSON
var DefaultOptions = Options{
	Endpoints: []string{"localhost:2379"},
}

// NewClient creates a new etcd client.
//
// You must call the Close() method on the client when you're done working with it.
func NewClient(options Options) (Store, error) {
	result := &Client{}

	// Set default values
	if options.Endpoints == nil || len(options.Endpoints) == 0 {
		options.Endpoints = DefaultOptions.Endpoints
	}

	config := clientv3.Config{
		Endpoints:   options.Endpoints,
		DialTimeout: 2 * time.Second,
		//DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}

	cli, err := clientv3.New(config)
	if err != nil {
		return result, err
	}

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	statusRes, err := cli.Status(ctxWithTimeout, options.Endpoints[0])
	if err != nil {
		return result, err
	} else if statusRes == nil {
		return result, errors.New("the status response from etcd was nil")
	}

	result.c = cli
	result.timeOut = defaultTimeout

	return result, nil
}
