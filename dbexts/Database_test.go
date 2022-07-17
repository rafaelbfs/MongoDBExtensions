package dbexts

import (
	"context"
	"fmt"
	a "github.com/rafaelbfs/GoConvenience/Assertions"
	convenience "github.com/rafaelbfs/GoConvenience/Convenience"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
)

const COLL_NAME = "PeopleTest"
const DB_NAME = "skillsdata"

type Person struct {
	ID        *primitive.ObjectID `bson:"_id,omitempty"`
	FirstName string              `bson:"first_name,omitempty"`
	LastName  string              `bson:"last_name,omitempty"`
}

func Setup(cfgfile string, dbname string) MongoCRUDRepo[Person] {
	Initialize(cfgfile, dbname)
	r := MongoCRUDRepo[Person]{}
	r.Init(COLL_NAME).Coll.Drop(context.TODO())
	r.Init(COLL_NAME)
	return r
}

func DeafaultSetup() MongoCRUDRepo[Person] {
	Initialize("../local.env", DB_NAME)
	r := MongoCRUDRepo[Person]{}
	r.Init(COLL_NAME).Coll.Drop(context.TODO())
	r.Init(COLL_NAME)
	return r
}

func mkTestPerson() Person {
	return Person{
		FirstName: "John",
		LastName:  "Doe",
	}
}

func TestRegex(t *testing.T) {
	str := "/var/environment.env"
	a.Assert(t).Condition(RegexEnvFile.MatchString(str)).IsTrueV()
	a.Assert(t).Condition(RegexEnvFile.MatchString("..local.env")).IsFalseV()
	a.Assert(t).Condition(RegexEnvFile.MatchString("../local.env")).IsTrueV()
	a.Assert(t).Condition(RegexEnvFile.MatchString("not.an.env.file")).IsFalseV()
	a.Assert(t).Condition(RegexEnvFile.MatchString("relativePath/local.env")).IsTrueV()
}

func TestPersonCRUD(t *testing.T) {
	Initialize("../local.env", DB_NAME)
	convenience.Nvl(GetDatabase().Collection(COLL_NAME)).DoIfPresent(
		func(c *mongo.Collection) { c.Drop(context.TODO()) })
	err := GetDatabase().CreateCollection(context.TODO(), COLL_NAME)
	a.Assert(t).NoError(err)

	coll := GetDatabase().Collection(COLL_NAME)
	a.AssertPointer(t, coll).NotNil()

	person := mkTestPerson()
	one, err := coll.InsertOne(context.TODO(), person)
	a.Assert(t).NoError(err)

	id := one.InsertedID.(primitive.ObjectID)
	fmt.Printf("Inserted ID=%v", id)
	a.Assert(t).Condition(id.IsZero()).IsFalseV()
	byId := bson.D{{"_id", id}}
	var bsonRes bson.Raw

	bsonRes, err = coll.FindOne(context.TODO(), byId).DecodeBytes()
	a.Assert(t).NoError(err)
	var entity Person
	err = bson.Unmarshal(bsonRes, &entity)
	a.Assert(t).NoError(err)
	t.Logf("Unmarshalled: %v", entity)
	a.AssertThat(t, entity.FirstName).EqualsTo(person.FirstName)
}
