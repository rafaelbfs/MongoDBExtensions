package dbexts

import (
	"fmt"
	a "github.com/rafaelbfs/GoConvenience/Assertions"
	convenience "github.com/rafaelbfs/GoConvenience/Convenience"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

func TestRetrieveById(t *testing.T) {
	r := DeafaultSetup()

	person := mkTestPerson()
	one := convenience.Try(r.Init(COLL_NAME).Create(person)).ResultOrPanic()
	a.Assert(t).Condition(one.InsertedID.(primitive.ObjectID).IsZero()).IsFalseV()

	id := one.InsertedID.(primitive.ObjectID).Hex()
	fmt.Printf("Inserted ID=%v", id)
	oid := convenience.Try(primitive.ObjectIDFromHex(id)).ResultOrPanic()

	ent := r.Init(COLL_NAME).RetrieveById(oid)
	a.AssertPointer(t, ent).NotNil()
	t.Logf("Unmarshalled: %v", ent)
	a.AssertThat(t, ent.FirstName).EqualsTo(person.FirstName)

	//Shutdown()
}

func TestMongoCRUDRepo_RetrieveByFilter(t *testing.T) {
	r := DeafaultSetup()

	john := Person{FirstName: "John", LastName: "Doe"}
	jane := Person{FirstName: "Jane", LastName: "Doe"}

	one, err := r.Create(jane)
	a.Assert(t).NoError(err)
	a.Assert(t).Condition(one.InsertedID.(primitive.ObjectID).IsZero()).IsFalseV()

	one, err = r.Create(john)
	a.Assert(t).NoError(err)
	a.Assert(t).Condition(one.InsertedID.(primitive.ObjectID).IsZero()).IsFalseV()

	ents, err := r.Init(COLL_NAME).RetrieveByFilter(&bson.D{{"last_name", "Doe"}})
	a.Assert(t).NoError(err)
	a.AssertThat(t, len(ents)).EqualsTo(2)

	//r.Init(COLL_NAME).Coll.Drop(context.TODO())
	//Shutdown()
}

func TestMongoCRUDRepo_Replace(t *testing.T) {
	r := DeafaultSetup()

	person := Person{FirstName: "John", LastName: "Doe"}

	one, err := r.Create(person)
	a.Assert(t).NoError(err)
	a.Assert(t).Condition(one.InsertedID.(primitive.ObjectID).IsZero()).IsFalseV()

	insertedPerson := r.RetrieveById(one.InsertedID.(primitive.ObjectID))

	a.AssertPointer(t, insertedPerson).NotNil()
	a.AssertThat(t, insertedPerson.FirstName).EqualsTo("John")
	a.Assert(t).Condition(insertedPerson.ID.IsZero()).IsFalseV()

	person.FirstName = "Josephine"

	oid := insertedPerson.ID
	fn := r.UpdateById(oid)
	updtate := MkUpdateSetStatement(*insertedPerson, person)
	res, err := fn(updtate)
	a.Assert(t).NoError(err)
	a.AssertThat(t, res.ModifiedCount).EqualsTo(1)

	//Shutdown()
}
