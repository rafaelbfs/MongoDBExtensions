package dbexts

import (
	"context"
	convenience "github.com/rafaelbfs/GoConvenience/Convenience"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
)

type MongoCRUD[A any] interface {
	Create(A) (*mongo.InsertOneResult, error)
	RetrieveById(oid primitive.ObjectID) *A
	RetrieveByFilter(filter *bson.D) ([]A, error)
	RetrieveWithLimit(filter *bson.D, limit int) ([]A, error)
	UpdateById(id *primitive.ObjectID) UpdateFn
	DeleteById(oid string) bool
}

type Initializable[A any] interface {
	Init() Initializable[A]
}

type MongoCRUDRepo[A any] struct {
	db   *mongo.Database
	Coll *mongo.Collection
}

func (m MongoCRUDRepo[A]) Create(b A) (*mongo.InsertOneResult, error) {
	return m.Coll.InsertOne(context.TODO(), b)
}

func (m MongoCRUDRepo[A]) RetrieveById(oid primitive.ObjectID) *A {
	var res A
	err := m.Coll.FindOne(context.TODO(), bson.D{{"_id", oid}}).Decode(&res)
	if err != nil {
		return nil
	}
	return &res
}

func (m MongoCRUDRepo[A]) RetrieveWithLimit(filter *bson.D, limit int) ([]A, error) {
	ctx := context.TODO()
	res, err := m.Coll.Find(ctx, filter)
	if err != nil {
		log.Printf("Suppressed error:%v", err)
		return nil, err
	}
	defer res.Close(ctx)

	if limit <= 0 {
		var all []A
		err = res.All(ctx, &all)
		return all, err
	}

	var list = make([]A, 0, limit)

	for res.Next(context.TODO()) {
		var result A
		if err = res.Decode(&result); err != nil {
			return list, err
		}
		list = append(list, result)
	}
	return list, res.Err()
}

func (m MongoCRUDRepo[A]) RetrieveByFilter(filter *bson.D) ([]A, error) {
	return m.RetrieveWithLimit(filter, -1)
}

func (m MongoCRUDRepo[A]) UpdateById(id *primitive.ObjectID) UpdateFn {
	if id == nil {
		return func(d bson.D) (*mongo.UpdateResult, error) {
			return nil, os.ErrInvalid
		}
	}
	return func(ops bson.D) (*mongo.UpdateResult, error) {
		return m.Coll.UpdateByID(context.TODO(), id, ops)
	}
}

func (m MongoCRUDRepo[A]) DeleteById(id string) int64 {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("\n Error %v. %v is not valid oid", err, id)
		return -1
	}
	r, err := m.Coll.DeleteOne(context.TODO(), bson.D{{"_id", oid}})
	if err != nil {
		log.Printf("Nothing deleted due to: %v", err)
		return -1
	}
	return r.DeletedCount
}

func CollFactory(collectionName string) func() *mongo.Collection {
	return func() *mongo.Collection {
		db := GetDatabase()
		coll := db.Collection(collectionName)
		if coll == nil {
			err := db.CreateCollection(context.TODO(), collectionName)
			convenience.WrapError(err).AndPanic()
			coll = db.Collection(collectionName)
		}
		return coll
	}
}

func (it *MongoCRUDRepo[A]) Init(collectionName string) *MongoCRUDRepo[A] {
	it.db = convenience.Nvl(it.db).OrCall(GetDatabase)
	it.Coll = convenience.Nvl(it.Coll).OrCall(CollFactory(collectionName))
	return it
}
