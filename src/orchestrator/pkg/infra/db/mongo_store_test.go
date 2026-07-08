/*
 * Copyright 2020 Intel Corporation, Inc
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
	"strings"
	"testing"

	"golang.org/x/net/context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/mongo/options"

	pkgerrors "github.com/pkg/errors"
)

// storeTestKey is a sample key used for the Store method tests.
type storeTestKey struct {
	Project string `json:"project"`
}

// mockCollectionCount embeds mockCollection but allows the CountDocuments
// return value to be customized (mockCollection always returns 1).
type mockCollectionCount struct {
	mockCollection
	count int64
}

func (c *mockCollectionCount) CountDocuments(ctx context.Context, filter interface{},
	opts ...*options.CountOptions) (int64, error) {
	return c.count, c.Err
}

// withMockedSeams swaps the module-level function seams (getCollection,
// decodeBytes, cursorNext, cursorClose) for the duration of fn and restores
// them afterwards.  This lets us drive the MongoStore methods without a live
// database.
func withMockedSeams(coll MongoCollection, decodeErr error, next bool, fn func()) {
	origGetCollection := getCollection
	origDecodeBytes := decodeBytes
	origCursorNext := cursorNext
	origCursorClose := cursorClose
	defer func() {
		getCollection = origGetCollection
		decodeBytes = origDecodeBytes
		cursorNext = origCursorNext
		cursorClose = origCursorClose
	}()

	getCollection = func(c string, m *MongoStore) MongoCollection {
		return coll
	}
	decodeBytes = func(sr *mongo.SingleResult) (bson.Raw, error) {
		return bson.Raw{}, decodeErr
	}
	// Return next only once (avoids infinite loops and nil cursor deref).
	called := false
	cursorNext = func(ctx context.Context, cursor *mongo.Cursor) bool {
		if called {
			return false
		}
		called = true
		return next
	}
	cursorClose = func(ctx context.Context, cursor *mongo.Cursor) error {
		return nil
	}

	fn()
}

func TestMongoStoreInsert(t *testing.T) {
	m := &MongoStore{}
	key := storeTestKey{Project: "p1"}

	t.Run("Missing data returns error", func(t *testing.T) {
		err := m.Insert("coll", key, nil, "tag", nil)
		if err == nil {
			t.Fatalf("Insert expected error for nil data")
		}
	})

	t.Run("Missing mandatory field returns error", func(t *testing.T) {
		err := m.Insert("", key, nil, "tag", "data")
		if err == nil {
			t.Fatalf("Insert expected error for empty collection name")
		}
	})

	t.Run("Successful insert without query", func(t *testing.T) {
		withMockedSeams(&mockCollection{}, nil, false, func() {
			err := m.Insert("coll", key, nil, "tag", "data")
			if err != nil {
				t.Fatalf("Insert returned an unexpected error: %s", err)
			}
		})
	})

	t.Run("Successful insert with query", func(t *testing.T) {
		withMockedSeams(&mockCollection{}, nil, false, func() {
			err := m.Insert("coll", key, storeTestKey{Project: "p2"}, "tag", "data")
			if err != nil {
				t.Fatalf("Insert returned an unexpected error: %s", err)
			}
		})
	})

	t.Run("FindOneAndUpdate error surfaces", func(t *testing.T) {
		withMockedSeams(&mockCollection{}, pkgerrors.New("decode failed"), false, func() {
			err := m.Insert("coll", key, nil, "tag", "data")
			if err == nil {
				t.Fatalf("Insert expected error when master table update fails")
			}
			if !strings.Contains(err.Error(), "Error updating master table") {
				t.Fatalf("Insert returned an unexpected error: %s", err)
			}
		})
	})
}

func TestMongoStoreFind(t *testing.T) {
	m := &MongoStore{}
	key := storeTestKey{Project: "p1"}

	t.Run("Missing mandatory field returns error", func(t *testing.T) {
		_, err := m.Find("", key, "tag")
		if err == nil {
			t.Fatalf("Find expected error for empty collection name")
		}
	})

	t.Run("Find error surfaces", func(t *testing.T) {
		withMockedSeams(&mockCollection{Err: pkgerrors.New("find failed")}, nil, false, func() {
			_, err := m.Find("coll", key, "tag")
			if err == nil {
				t.Fatalf("Find expected error when collection Find fails")
			}
			if !strings.Contains(err.Error(), "Error finding element") {
				t.Fatalf("Find returned an unexpected error: %s", err)
			}
		})
	})

	t.Run("Find with no documents returns empty", func(t *testing.T) {
		withMockedSeams(&mockCollection{}, nil, false, func() {
			result, err := m.Find("coll", key, "tag")
			if err != nil {
				t.Fatalf("Find returned an unexpected error: %s", err)
			}
			if len(result) != 0 {
				t.Fatalf("Find expected empty result, got %v", result)
			}
		})
	})
}

func TestMongoStoreRemove(t *testing.T) {
	m := &MongoStore{}
	key := storeTestKey{Project: "p1"}

	t.Run("Missing mandatory field returns error", func(t *testing.T) {
		err := m.Remove("", key)
		if err == nil {
			t.Fatalf("Remove expected error for empty collection name")
		}
	})

	t.Run("CountDocuments greater than one blocks delete", func(t *testing.T) {
		// mockCollection.CountDocuments returns 1 by default, so inject a
		// custom collection to exercise the child-reference guard.
		withMockedSeams(&mockCollectionCount{count: 2}, nil, false, func() {
			err := m.Remove("coll", key)
			if err == nil {
				t.Fatalf("Remove expected error when child references exist")
			}
			if !strings.Contains(err.Error(), "child references") {
				t.Fatalf("Remove returned an unexpected error: %s", err)
			}
		})
	})

	t.Run("Successful remove", func(t *testing.T) {
		withMockedSeams(&mockCollectionCount{count: 1}, nil, false, func() {
			err := m.Remove("coll", key)
			if err != nil {
				t.Fatalf("Remove returned an unexpected error: %s", err)
			}
		})
	})
}

func TestMongoStoreRemoveAll(t *testing.T) {
	m := &MongoStore{}
	key := storeTestKey{Project: "p1"}

	t.Run("Missing mandatory field returns error", func(t *testing.T) {
		err := m.RemoveAll("", key)
		if err == nil {
			t.Fatalf("RemoveAll expected error for empty collection name")
		}
	})

	t.Run("Successful remove all", func(t *testing.T) {
		withMockedSeams(&mockCollection{}, nil, false, func() {
			err := m.RemoveAll("coll", key)
			if err != nil {
				t.Fatalf("RemoveAll returned an unexpected error: %s", err)
			}
		})
	})

	t.Run("DeleteMany error surfaces", func(t *testing.T) {
		withMockedSeams(&mockCollection{Err: pkgerrors.New("delete failed")}, nil, false, func() {
			err := m.RemoveAll("coll", key)
			if err == nil {
				t.Fatalf("RemoveAll expected error when DeleteMany fails")
			}
		})
	})
}

func TestMongoStoreRemoveTag(t *testing.T) {
	m := &MongoStore{}
	key := storeTestKey{Project: "p1"}

	t.Run("Successful remove tag", func(t *testing.T) {
		withMockedSeams(&mockCollection{}, nil, false, func() {
			err := m.RemoveTag("coll", key, "tag")
			if err != nil {
				t.Fatalf("RemoveTag returned an unexpected error: %s", err)
			}
		})
	})

	t.Run("FindOneAndUpdate error surfaces", func(t *testing.T) {
		withMockedSeams(&mockCollection{}, pkgerrors.New("decode failed"), false, func() {
			err := m.RemoveTag("coll", key, "tag")
			if err == nil {
				t.Fatalf("RemoveTag expected error when FindOneAndUpdate fails")
			}
			if !strings.Contains(err.Error(), "Error removing tag") {
				t.Fatalf("RemoveTag returned an unexpected error: %s", err)
			}
		})
	})
}
