package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clients map[string]*mongo.Client
	mutex   *sync.Mutex
)

func init() {
	clients = map[string]*mongo.Client{}
	mutex = &sync.Mutex{}
}

// I to get instance of database object.
func I(user string, password string, host string) *mongo.Client {
	key := host

	if val, ok := clients[key]; ok {
		return val
	}

	mutex.Lock()

	if val, ok := clients[key]; ok {
		return val
	}

	port := "27017"
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		host = parts[0]
		port = parts[1]
	}

	uri := "mongodb://"
	if user != "" && password != "" {
		uri = fmt.Sprintf("%s%s:%s@", uri, user, password)
	}
	uri = fmt.Sprintf("%s%s:%s", uri, host, port)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}

	clients[key] = client

	mutex.Unlock()

	return client
}
