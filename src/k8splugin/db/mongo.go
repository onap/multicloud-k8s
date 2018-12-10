/*
 * Copyright 2018 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package db

import (
	uuid "github.com/hashicorp/go-uuid"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	pkgerrors "github.com/pkg/errors"
	"golang.org/x/net/context"
	"log"
	"os"
)

// MongoCollection defines the a subset of MongoDB operations
// Note: This interface is defined mainly for mock testing
type MongoCollection interface {
	InsertOne(ctx context.Context, document interface{},
		opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{},
		opts ...*options.FindOneOptions) *mongo.SingleResult
	FindOneAndUpdate(ctx context.Context, filter interface{},
		update interface{}, opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
	DeleteOne(ctx context.Context, filter interface{},
		opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	Find(ctx context.Context, filter interface{},
		opts ...*options.FindOptions) (mongo.Cursor, error)
}

// MongoDatabase defines the a subset of MongoDB operations
// Note: This interface is defined mainly for mock testing
type MongoDatabase interface {
	Collection(name string, opts ...*options.CollectionOptions) *mongo.Collection
}

// MongoStore is an implementation of the MongoDatabase interface
type MongoStore struct {
	db MongoDatabase
}

// NewMongoStore initializes a Mongo Database with the name provided
// If a database with that name exists, it will be returned
func NewMongoStore(name string, store MongoDatabase) (Store, error) {
	if store == nil {
		ip := "mongodb://" + os.Getenv("DATABASE_IP") + ":27017"
		mongoClient, err := mongo.NewClient(ip)
		if err != nil {
			return nil, err
		}

		store = mongoClient.Database(name)
	}

	return &MongoStore{
		db: store,
	}, nil
}

// HealthCheck verifies if the database is up and running
func (m *MongoStore) HealthCheck() error {
	coll := "healthcheck"

	testID, _ := uuid.GenerateUUID()
	hcTag := "healthcheck"
	err := m.Create(coll, testID, hcTag, bson.D{{"test", "healthcheck"}})
	if err != nil {
		return pkgerrors.New("Unable to create healtcheck entry in database")
	}

	err = m.Delete(coll, testID, hcTag)
	if err != nil {
		return pkgerrors.New("Unable to delete healtcheck entry in database")
	}
	return nil
}

// Create is used to create a DB entry
func (m *MongoStore) Create(coll, key, tag string, data interface{}) error {
	if data == nil || key == "" || tag == "" || coll == "" {
		return pkgerrors.New("No Data to store")
	}

	c := m.db.Collection(coll)

	//Insert the data and then add the objectID to the masterTable
	res, err := c.InsertOne(context.Background(), bson.D{
		{tag, data},
	})
	if err != nil {
		return pkgerrors.Errorf("Error inserting into database: %s", err.Error())
	}

	//Add objectID of created data to masterKey document
	//Create masterkey document if it does not exist
	filter := bson.D{{"key", key}}
	c.FindOneAndUpdate(context.Background(), filter, bson.D{
		{"$set", bson.D{
			{tag, res.InsertedID},
		}},
	}, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After))

	return nil
}

// Read method returns the data stored for this key and for this particular tag
func (m *MongoStore) Read(coll, key, tag string) ([]byte, error) {
	if key == "" || tag == "" || coll == "" {
		return nil, pkgerrors.New("Mandatory fields are missing")
	}

	c := m.db.Collection(coll)
	ctx := context.Background()

	//Get the masterkey document based on given key
	filter := bson.D{{"key", key}}
	keydata := bson.Raw{}
	err := c.FindOne(context.Background(), filter).Decode(&keydata)
	if err != nil {
		return nil, pkgerrors.Errorf("Error finding master table: %s", err.Error())
	}

	//Read the tag objectID from document
	tagoid, ok := keydata.Lookup(tag).ObjectIDOK()
	if !ok {
		return nil, pkgerrors.Errorf("Error finding objectID for tag %s", tag)
	}

	//Use tag objectID to read the data from store
	filter = bson.D{{"_id", tagoid}}
	tagdata := bson.Raw{}
	err = c.FindOne(ctx, filter).Decode(&tagdata)
	if err != nil {
		return nil, pkgerrors.Errorf("Error reading found object: %s", err.Error())
	}

	//Return the data as a byte array
	return tagdata.Lookup(tag).Value, nil
}

// Helper function that deletes an object by its ID
func (m *MongoStore) deleteObjectByID(coll string, objID primitive.ObjectID) error {
	if coll == "" {
		return pkgerrors.New("Mandatory fields are missing")
	}

	c := m.db.Collection(coll)
	ctx := context.Background()

	_, err := c.DeleteOne(ctx, bson.D{{"_id", objID}})
	if err != nil {
		return pkgerrors.Errorf("Error Deleting from database: %s", err.Error())
	}

	log.Printf("Deleted Obj with ID %s", objID.String())
	return nil
}

// Delete method removes a document from the Database that matches key
func (m *MongoStore) Delete(coll, key, tag string) error {
	if key == "" || tag == "" || coll == "" {
		return pkgerrors.New("Mandatory fields are missing")
	}

	c := m.db.Collection(coll)
	ctx := context.Background()

	//Get the masterkey document based on given key
	filter := bson.D{{"key", key}}
	//Remove the tag ID entry from masterkey table
	update := bson.D{
		{
			"$unset", bson.D{
				{tag, ""},
			},
		},
	}
	keydata := bson.Raw{}
	err := c.FindOneAndUpdate(ctx, filter, update,
		options.FindOneAndUpdate().SetReturnDocument(options.Before)).Decode(&keydata)
	if err != nil {
		return pkgerrors.Errorf("Error decoding master table after update: %s",
			err.Error())
	}

	//Read the tag objectID from document
	tagoid, ok := keydata.Lookup(tag).ObjectIDOK()
	if !ok {
		return pkgerrors.Errorf("Error finding objectID for tag %s", tag)
	}

	//Use tag objectID to read the data from store
	err = m.deleteObjectByID(coll, tagoid)
	if err != nil {
		return pkgerrors.Errorf("Error deleting from database: %s", err.Error())
	}

	return nil
}

// ReadAll is used to get all documents in db of a particular tag
func (m *MongoStore) ReadAll(coll, tag string) (map[string][]byte, error) {
	if coll == "" || tag == "" {
		return nil, pkgerrors.New("Missing collection or tag name")
	}

	c := m.db.Collection(coll)
	ctx := context.Background()

	//Get all master tables in this collection
	filter := bson.D{
		{"key", bson.D{
			{"$exists", true},
		}},
	}
	cursor, err := c.Find(ctx, filter)
	if err != nil {
		return nil, pkgerrors.Errorf("Error reading from database: %s", err.Error())
	}
	defer cursor.Close(ctx)

	//Iterate over all the master tables
	var result map[string][]byte
	for cursor.Next(ctx) {
		d, err := cursor.DecodeBytes()
		if err != nil {
			log.Printf("Unable to decode data in Readall: %s", err.Error())
			continue
		}

		//Read key of each master table
		key, ok := d.Lookup("key").StringValueOK()
		if !ok {
			log.Printf("Unable to read key string from mastertable %s", err.Error())
			continue
		}

		//Get objectID of tag document
		tid, ok := d.Lookup(tag).ObjectIDOK()
		if !ok {
			log.Printf("Did not find tag: %s", tag)
			continue
		}

		//Find tag document and unmarshal it into []byte
		tagData := bson.Raw{}
		err = c.FindOne(ctx, bson.D{{"_id", tid}}).Decode(&tagData)
		if err != nil {
			log.Printf("Unable to decode tag data %s", err.Error())
			continue
		}
		result[key] = tagData.Lookup(tag).Value
	}

	return result, nil
}
