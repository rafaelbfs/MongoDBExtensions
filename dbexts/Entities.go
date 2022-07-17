package dbexts

import (
	convenience "github.com/rafaelbfs/GoConvenience/Convenience"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"regexp"
	"strings"
)

const (
	SET  = "$set"
	BSON = "bson"
	OID  = "_id"
)

var HexRegex *regexp.Regexp

type ReflectiveGet[C comparable] func(v reflect.Value) C

func init() {
	HexRegex = regexp.MustCompilePOSIX("^(0x|0X)?[a-fA-F0-9]+$")
}

type Entity struct {
}

func returnNil(err error) *primitive.ObjectID {
	return nil
}

type UpdateFn func(d bson.D) (*mongo.UpdateResult, error)

func ToOID(id string) *primitive.ObjectID {
	if !HexRegex.MatchString(id) {
		return nil
	}
	r, err := primitive.ObjectIDFromHex(id)
	return convenience.Try(&r, err).HandleErr(returnNil)
}

func emptyD() bson.D {
	return bson.D{}
}

func ExtractValue(v reflect.Value) interface{} {
	switch v.Type().Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return v.Int()
	case reflect.Float64, reflect.Float32:
		return v.Float()
	}
	return v.String()
}

func GetUpdate(ty reflect.Type, fieldName string, v1 reflect.Value, v2 reflect.Value) *bson.E {
	if !ty.Comparable() || reflect.DeepEqual(v1.Interface(), v2.Interface()) {
		return nil
	}
	return &bson.E{Key: fieldName, Value: ExtractValue(v2)}
}

func MakeUpdateStatements[A any](o A, n A) bson.D {
	if reflect.TypeOf(o).Kind() != reflect.Struct {
		return emptyD()
	}
	t := reflect.TypeOf(o)
	updates := make(bson.D, 0, t.NumField()-1)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.Type.Comparable() {
			continue
		}
		var name = field.Name

		if str, ok := field.Tag.Lookup(BSON); ok {
			if strings.Contains(str, OID) {
				continue
			}
			elements := strings.Split(str, ",")

			if len(elements) > 0 {
				name = elements[0]
			}
		}

		v := GetUpdate(field.Type, name, reflect.ValueOf(o).Field(i), reflect.ValueOf(n).Field(i))
		if v != nil {
			updates = append(updates, *v)
		}
	}
	return updates
}

func MkUpdateSetStatement[A any](o A, n A) bson.D {
	return bson.D{{SET, MakeUpdateStatements(o, n)}}
}
