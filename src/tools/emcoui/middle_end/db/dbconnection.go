/*
=======================================================================
Copyright (c) 2017-2020 Aarna Networks, Inc.
All rights reserved.
======================================================================
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
          http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
========================================================================
*/

package db

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

// MongoStore is the interface which implements the db.Store interface
type MongoStore struct {
	db *mongo.Database
}

// Key interface
type Key interface {
}

// DBconn variable of type Store
var DBconn Store

// Store Interface which implements the data store functions
type Store interface {
	HealthCheck() error
	Find(coll string, key []byte, tag string) ([][]byte, error)
	Unmarshal(inp []byte, out interface{}) error
}

// NewMongoStore Return mongo client
func NewMongoStore(name string, store *mongo.Database, svcEp string) (Store, error) {
	if store == nil {
		ip := "mongodb://" + svcEp
		clientOptions := options.Client()
		clientOptions.ApplyURI(ip)
		mongoClient, err := mongo.NewClient(clientOptions)
		if err != nil {
			return nil, err
		}

		err = mongoClient.Connect(context.Background())
		if err != nil {
			return nil, err
		}
		store = mongoClient.Database(name)
	}
	return &MongoStore{
		db: store,
	}, nil
}

// CreateDBClient creates the DB client. currently only mongo
func CreateDBClient(dbType string, dbName string, svcEp string) error {
	var err error
	switch dbType {
	case "mongo":
		DBconn, err = NewMongoStore(dbName, nil, svcEp)
	default:
		fmt.Println(dbType + "DB not supported")
	}
	return err
}

// HealthCheck verifies the database connection
func (m *MongoStore) HealthCheck() error {
	_, err := (*mongo.SingleResult).DecodeBytes(m.db.RunCommand(context.Background(), bson.D{{"serverStatus", 1}}))
	if err != nil {
		fmt.Println("Error getting DB server status: err %s", err)
	}
	return nil
}

func (m *MongoStore) Unmarshal(inp []byte, out interface{}) error {
	err := bson.Unmarshal(inp, out)
	if err != nil {
		fmt.Printf("Failed to unmarshall bson")
		return err
	}
	return nil
}

// Find a document
func (m *MongoStore) Find(coll string, key []byte, tag string) ([][]byte, error) {
	var bsonMap bson.M
	err := json.Unmarshal([]byte(key), &bsonMap)
	if err != nil {
		fmt.Println("Failed to unmarshall %s\n", key)
		return nil, err
	}

	filter := bson.M{
		"$and": []bson.M{bsonMap},
	}

	fmt.Printf("%+v %s\n", filter, tag)
	projection := bson.D{
		{tag, 1},
		{"_id", 0},
	}

	c := m.db.Collection(coll)

	cursor, err := c.Find(context.Background(), filter, options.Find().SetProjection(projection))
	if err != nil {
		fmt.Println("Failed to find the document %s\n", err)
		return nil, err
	}

	defer cursor.Close(context.Background())
	var data []byte
	var result [][]byte
	for cursor.Next(context.Background()) {
		d := cursor.Current
		switch d.Lookup(tag).Type {
		case bson.TypeString:
			data = []byte(d.Lookup(tag).StringValue())
		default:
			r, err := d.LookupErr(tag)
			if err != nil {
				fmt.Println("Unable to read data %s %s\n", string(r.Value), err)
			}
			data = r.Value
		}
		result = append(result, data)
	}
	return result, nil
}
