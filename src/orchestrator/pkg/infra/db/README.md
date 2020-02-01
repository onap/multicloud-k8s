# Database Abstraction Layer

This package contains implementations of the Database interface defined in `store.go`
Any database can be used as the backend as long as the following interface is implemented;

```go
type Store interface {
	// Returns nil if db health is good
	HealthCheck() error

	// Unmarshal implements any unmarshaling needed for the database
	Unmarshal(inp []byte, out interface{}) error

	// Creates a new master table with key and links data with tag and
	// creates a pointer to the newly added data in the master table
	Create(table string, key Key, tag string, data interface{}) error

	// Reads data for a particular key with specific tag.
	Read(table string, key Key, tag string) ([]byte, error)

	// Update data for particular key with specific tag
	Update(table string, key Key, tag string, data interface{}) error

	// Deletes a specific tag data for key.
	// TODO: If tag is empty, it will delete all tags under key.
	Delete(table string, key Key, tag string) error

	// Reads all master tables and data from the specified tag in table
	ReadAll(table string, tag string) (map[string][]byte, error)
}
```

Therefore, `mongo.go`, `consul.go` implement the above interface and can be used as the backend as needed based on initial configuration.

## Details on Mongo Implementation

`mongo.go` implements the above interface using the `go.mongodb.org/mongo-driver` package.
The code converts incoming binary data and creates a new document in the database.

### Create

Arguments:
```go
collection string
key interface
tag string
data []byte
```

Create inserts the provided `data` into the `collection` which returns an auto-generated (by `mongodb`) ID which we then associate with the `key` that is provided as one of the arguments.

We use the `FindOneAndUpdate` mongo API to achieve this with the `upsert` option set to `true`.
We create the following documents in mongodb for each new definition added to the database:

There is a Master Key document that contains references to other documents which are related to this `key`.

#### Master Key Entry
```json
{
    "_id" : ObjectId("5e0a8554b78a15f71d2dce7e"),
    "key" : { "rbname" : "edgex", "rbversion" : "v1"},
    "defmetadata" : ObjectId("5e0a8554be261ecb57f067eb"),
    "defcontent" : ObjectId("5e0a8377bcfcdd0f01dc7b0d")
}
```
#### Metadata Key Entry
```json
{
    "_id" : ObjectId("5e0a8554be261ecb57f067eb"),
    "defmetadata" : { "rbname" : "edgex", "rbversion" : "v1", "chartname" : "", "description" : "", "labels" : null }
}
```
#### Definition Content
```json
{
    "_id" : ObjectId("5e0a8377bcfcdd0f01dc7b0d"),
    "defcontent" : "H4sICCVd3FwAA3Byb2ZpbGUxLnRhcgDt1NEKgjAUxvFd7ylG98aWOsGXiYELxLRwJvj2rbyoIPDGiuD/uzmwM9iB7Vvruvrgw7CdXHsUn6Ejm2W3aopcP9eZLYRJM1voPN+ZndAm16kVSn9onheXMLheKeGqfdM0rq07/3bfUv9PJUkiR9+H+tSVajRymM6+lEqN7njxoVSbU+z2deX388r9nWzkr8fGSt5d79pnLOZfm0f+dRrzb7P4DZD/LyDJAAAAAAAAAAAAAAAA/+0Ksq1N5QAoAAA="
}
```

### Unmarshal

Data in mongo is stored as `bson` which is a compressed form of `json`. We need mongo to convert the stored `bson` data to regular `json`
that we can use in our code when returned.

We just use the `bson.Unmarshal` API to achieve this.

### Read

Arguments:
```go
collection string
key interface
tag string
```

Read is straight forward and it uses the `FindOne` API to find our Mongo document based on the provided `key` and then gets the corresponding data for the given `tag`. It will return []byte which can then be passed to the `Unmarshal` function to get the desired GO object.

### Delete

Delete is similar to Read and deletes all the objectIDs being stored for a given `key` in the collection.

## Testing Interfaces

The following interface exists to allow for the development of unit tests which don't require mongo to be running.
It is mentioned so in the code as well.

```go
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
```