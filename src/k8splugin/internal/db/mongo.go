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
	"golang.org/x/net/context"
	"log"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"

	pkgerrors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		opts ...*options.FindOptions) (*mongo.Cursor, error)
}

// MongoStore is an implementation of the db.Store interface
type MongoStore struct {
	db *mongo.Database
}

// This exists only for allowing us to mock the collection object
// for testing purposes
var getCollection = func(coll string, m *MongoStore) MongoCollection {
	return m.db.Collection(coll)
}

// This exists only for allowing us to mock the DecodeBytes function
// Mainly because we cannot construct a SingleResult struct from our
// tests. All fields in that struct are private.
var decodeBytes = func(sr *mongo.SingleResult) (bson.Raw, error) {
	return sr.DecodeBytes()
}

// These exists only for allowing us to mock the cursor.Next function
// Mainly because we cannot construct a mongo.Cursor struct from our
// tests. All fields in that struct are private and there is no public
// constructor method.
var cursorNext = func(ctx context.Context, cursor *mongo.Cursor) bool {
	return cursor.Next(ctx)
}
var cursorClose = func(ctx context.Context, cursor *mongo.Cursor) error {
	return cursor.Close(ctx)
}

// NewMongoStore initializes a Mongo Database with the name provided
// If a database with that name exists, it will be returned
func NewMongoStore(name string, store *mongo.Database) (Store, error) {
	if store == nil {
		ip := "mongodb://" + config.GetConfiguration().DatabaseAddress + ":27017"
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

// HealthCheck verifies if the database is up and running
func (m *MongoStore) HealthCheck() error {

	_, err := decodeBytes(m.db.RunCommand(context.Background(), bson.D{{"serverStatus", 1}}))
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting server status")
	}

	return nil
}

// validateParams checks to see if any parameters are empty
func (m *MongoStore) validateParams(args ...interface{}) bool {
	for _, v := range args {
		val, ok := v.(string)
		if ok {
			if val == "" {
				return false
			}
		} else {
			if v == nil {
				return false
			}
		}
	}

	return true
}

// Create is used to create a DB entry
func (m *MongoStore) Create(coll string, key Key, tag string, data interface{}) error {
	if data == nil || !m.validateParams(coll, key, tag) {
		return pkgerrors.New("No Data to store")
	}

	c := getCollection(coll, m)
	ctx := context.Background()

	//Insert the data and then add the objectID to the masterTable
	res, err := c.InsertOne(ctx, bson.D{
		{tag, data},
	})
	if err != nil {
		return pkgerrors.Errorf("Error inserting into database: %s", err.Error())
	}

	//Add objectID of created data to masterKey document
	//Create masterkey document if it does not exist
	filter := bson.D{{"key", key}}

	_, err = decodeBytes(
		c.FindOneAndUpdate(
			ctx,
			filter,
			bson.D{
				{"$set", bson.D{
					{tag, res.InsertedID},
				}},
			},
			options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)))

	if err != nil {
		return pkgerrors.Errorf("Error updating master table: %s", err.Error())
	}

	return nil
}

// Update is used to update a DB entry
func (m *MongoStore) Update(coll string, key Key, tag string, data interface{}) error {
	if data == nil || !m.validateParams(coll, key, tag) {
		return pkgerrors.New("No Data to update")
	}

	c := getCollection(coll, m)
	ctx := context.Background()

	//Get the masterkey document based on given key
	filter := bson.D{{"key", key}}
	keydata, err := decodeBytes(c.FindOne(context.Background(), filter))
	if err != nil {
		return pkgerrors.Errorf("Error finding master table: %s", err.Error())
	}

	//Read the tag objectID from document
	tagoid, ok := keydata.Lookup(tag).ObjectIDOK()
	if !ok {
		return pkgerrors.Errorf("Error finding objectID for tag %s", tag)
	}

	//Update the document with new data
	filter = bson.D{{"_id", tagoid}}

	_, err = decodeBytes(
		c.FindOneAndUpdate(
			ctx,
			filter,
			bson.D{
				{"$set", bson.D{
					{tag, data},
				}},
			},
			options.FindOneAndUpdate().SetReturnDocument(options.After)))

	if err != nil {
		return pkgerrors.Errorf("Error updating record: %s", err.Error())
	}

	return nil
}

// Unmarshal implements an unmarshaler for bson data that
// is produced from the mongo database
func (m *MongoStore) Unmarshal(inp []byte, out interface{}) error {
	err := bson.Unmarshal(inp, out)
	if err != nil {
		return pkgerrors.Wrap(err, "Unmarshaling bson")
	}
	return nil
}

// Read method returns the data stored for this key and for this particular tag
func (m *MongoStore) Read(coll string, key Key, tag string) ([]byte, error) {
	if !m.validateParams(coll, key, tag) {
		return nil, pkgerrors.New("Mandatory fields are missing")
	}

	c := getCollection(coll, m)
	ctx := context.Background()

	//Get the masterkey document based on given key
	filter := bson.D{{"key", key}}
	keydata, err := decodeBytes(c.FindOne(context.Background(), filter))
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
	tagdata, err := decodeBytes(c.FindOne(ctx, filter))
	if err != nil {
		return nil, pkgerrors.Errorf("Error reading found object: %s", err.Error())
	}

	//Return the data as a byte array
	//Convert string data to byte array using the built-in functions
	switch tagdata.Lookup(tag).Type {
	case bson.TypeString:
		return []byte(tagdata.Lookup(tag).StringValue()), nil
	default:
		return tagdata.Lookup(tag).Value, nil
	}
}

// Helper function that deletes an object by its ID
func (m *MongoStore) deleteObjectByID(coll string, objID primitive.ObjectID) error {

	c := getCollection(coll, m)
	ctx := context.Background()

	_, err := c.DeleteOne(ctx, bson.D{{"_id", objID}})
	if err != nil {
		return pkgerrors.Errorf("Error Deleting from database: %s", err.Error())
	}

	log.Printf("Deleted Obj with ID %s", objID.String())
	return nil
}

// Delete method removes a document from the Database that matches key
// TODO: delete all referenced docs if tag is empty string
func (m *MongoStore) Delete(coll string, key Key, tag string) error {
	if !m.validateParams(coll, key, tag) {
		return pkgerrors.New("Mandatory fields are missing")
	}

	c := getCollection(coll, m)
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
	keydata, err := decodeBytes(c.FindOneAndUpdate(ctx, filter, update,
		options.FindOneAndUpdate().SetReturnDocument(options.Before)))
	if err != nil {
		//No document was found. Return nil.
		if err == mongo.ErrNoDocuments {
			return nil
		}
		//Return any other error that was found.
		return pkgerrors.Errorf("Error decoding master table after update: %s",
			err.Error())
	}

	//Read the tag objectID from document
	elems, err := keydata.Elements()
	if err != nil {
		return pkgerrors.Errorf("Error reading elements from database: %s", err.Error())
	}

	tagoid, ok := keydata.Lookup(tag).ObjectIDOK()
	if !ok {
		return pkgerrors.Errorf("Error finding objectID for tag %s", tag)
	}

	//Use tag objectID to read the data from store
	err = m.deleteObjectByID(coll, tagoid)
	if err != nil {
		return pkgerrors.Errorf("Error deleting from database: %s", err.Error())
	}

	//Delete master table if no more tags left
	//_id, key and tag should be elements in before doc
	//if master table needs to be removed too
	if len(elems) == 3 {
		keyid, ok := keydata.Lookup("_id").ObjectIDOK()
		if !ok {
			return pkgerrors.Errorf("Error finding objectID for key %s", key)
		}
		err = m.deleteObjectByID(coll, keyid)
		if err != nil {
			return pkgerrors.Errorf("Error deleting master table from database: %s", err.Error())
		}
	}

	return nil
}

// ReadAll is used to get all documents in db of a particular tag
func (m *MongoStore) ReadAll(coll, tag string) (map[string][]byte, error) {
	if !m.validateParams(coll, tag) {
		return nil, pkgerrors.New("Missing collection or tag name")
	}

	c := getCollection(coll, m)
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
	defer cursorClose(ctx, cursor)

	//Iterate over all the master tables
	result := make(map[string][]byte)
	for cursorNext(ctx, cursor) {
		d := cursor.Current

		//Read key of each master table
		key, ok := d.Lookup("key").DocumentOK()
		if !ok {
			//Throw error if key is not found
			pkgerrors.New("Unable to read key from mastertable")
		}

		//Get objectID of tag document
		tid, ok := d.Lookup(tag).ObjectIDOK()
		if !ok {
			log.Printf("Did not find tag: %s", tag)
			continue
		}

		//Find tag document and unmarshal it into []byte
		tagData, err := decodeBytes(c.FindOne(ctx, bson.D{{"_id", tid}}))
		if err != nil {
			log.Printf("Unable to decode tag data %s", err.Error())
			continue
		}
		result[key.String()] = tagData.Lookup(tag).Value
	}

	if len(result) == 0 {
		return result, pkgerrors.Errorf("Did not find any objects with tag: %s", tag)
	}

	return result, nil
}
