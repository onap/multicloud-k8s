# Database Abstraction Layer

This package contains implementations of the Database interface defined in `store.go`
Any database can be used as the backend as long as the following interface is implemented;

```go
type Store interface {
	// Returns nil if db health is good
	HealthCheck() error

	// Unmarshal implements any unmarshaling needed for the database
	Unmarshal(inp []byte, out interface{}) error

	// Inserts and Updates a tag with key and also adds query fields if provided
	Insert(coll string, key Key, query interface{}, tag string, data interface{}) error

	// Find the document(s) with key and get the tag values from the document(s)
	Find(coll string, key Key, tag string) ([][]byte, error)

	// Removes the document(s) matching the key
	Remove(coll string, key Key) error
}
```

With this interface multiple database types can be supported by providing backends.

## Details on Mongo Implementation

`mongo.go` implements the above interface using the `go.mongodb.org/mongo-driver` package.
The code converts incoming binary data and creates a new document in the database.

### Insert

Arguments:
```go
collection string
key interface
query interface
tag string
data []byte
```

Insert function inserts the provided `data` into the `collection` as a document in MongoDB. `FindOneAndUpdate` mongo API is used to achieve this with the `upsert` option set to `true`. With this if the record doesn't exist it is created and if it exists it is updated with new data for the tag.

Key and Query parameters are assumed to be json structures with each element as part of the key. Those key-value pairs are used as the key for the document.
Internally this API takes all the fields in the Key structure and adds them as fields in the document. Query parameter works just like key and it is used to add additional fields to the document. 

With this key the document can be quried with Mongo `Find` function for both the key fields and Query fields.

This API also adds another field called "Key" field to the document. The "Key" field is concatenation of the key part of the Key parameter. Internally this is used to figure out the type of the document.

Assumption is that all the elememts of the key structure are strings.

#### Example of Key Structure
```go
type CompositeAppKey struct {
	CompositeAppName string `json:"compositeappname"`
	Version          string `json:"version"`
	Project          string `json:"project"`
}
```
#### Example of Key Values
```go
key := CompositeAppKey{
		CompositeAppName:  "ca1",
		Version:           "v1",
		Project:           "testProject",
	}
```

#### Example of Query Structure
```go
type Query struct {
	Userdata1 string `json:"userdata1"`
}
```
#### Example of Document store in MongoDB
```json
{
   "_id":"ObjectId("   "5e54c206f53ca130893c8020"   ")",
   "compositeappname":"ca1",
   "project":"testProject",
   "version":"v1",
   "compositeAppmetadata":{
      "metadata":{
         "name":"ca1",
         "description":"Test ca",
         "userdata1":"Data1",
         "userdata2":"Data2"
      },
      "spec":{
         "version":"v1"
      }
   },
   "key":"{compositeappname,project,version,}"
}
```

### Find

Arguments:
```go
collection string
key interface
tag string
```

Find function return one or more tag data based on the Key value. If key has all the fields defined then an exact match is looked for based on the key passed in. 
If some of the field value in structure are empty strings then this function returns all the documents which have the same type. (ANY operation)

#### Example of Exact Match based on fields Key Values
```go
key := CompositeAppKey{
		CompositeAppName:  "ca1",
		Version:           "v1",
		Project:           "testProject",
	}
```

#### Example of Match based on some fields
This example will return all the compositeApp under project testProject.
```go
key := CompositeAppKey{
		Project:           "testProject",
		CompositeAppName:  "",
		Version:           "",
		
	}
```

NOTE: Key structure can be different from the original key and can include Query fields also. ANY operation is not supported for Query fields.

### RemoveAll

Arguments:
```go
collection string
key interface
```
Similar to find. This will remove one or more documents based on the key structure.

### Remove

Arguments:
```go
collection string
key interface
```
This will remove one document based on the key structure. If child refrences exist for the key then the document will not be removed.

### Unmarshal

Data in mongo is stored as `bson` which is a compressed form of `json`. We need mongo to convert the stored `bson` data to regular `json`
that we can use in our code when returned.

`bson.Unmarshal` API is used to achieve this.



