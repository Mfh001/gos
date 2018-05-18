#!/usr/bin/env bash
basePath=$(pwd)/vendor

mkdir -p $basePath
export GOPATH=$basePath

# Redis
go get -u github.com/go-redis/redis

# Iris
go get -u github.com/kataras/iris
go get -u github.com/hashicorp

#gRpc
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
go get -u google.golang.org/grpc

#BDD
go get -u github.com/onsi/ginkgo/ginkgo
go get -u github.com/onsi/gomega/...

#Skiplist
go get -u github.com/ryszard/goskiplist/skiplist

#ORM
go get -u github.com/go-gorp/gorp
go get -u github.com/go-sql-driver/mysql

#UUID
go get -u github.com/rs/xid

#Cron
go get -u github.com/robfig/cron

#iOS Push
go get -u github.com/sideshow/apns2

#Android FCM
go get -u github.com/appleboy/go-fcm

#In App Purchase Validator
go get -u github.com/awa/go-iap/appstore
go get -u github.com/awa/go-iap/playstore