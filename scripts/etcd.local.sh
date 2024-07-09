#!/bin/bash

docker run \
    --publish 2379:2379 \
    --publish 4001:4001 \
    --name etcd \
    --env ALLOW_NONE_AUTHENTICATION=yes \
    --env ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379,http://0.0.0.0:4001 \
    bitnami/etcd:latest
