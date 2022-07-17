package dbexts

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	conv "github.com/rafaelbfs/GoConvenience/Convenience"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"regexp"
)

var (
	env           map[string]string
	client        *mongo.Client
	database      *mongo.Database
	RegexEnvFile  *regexp.Regexp
	RegexDatabase *regexp.Regexp
	uri           string
	databaseName  string
)

func resolveFile(st string) string {
	if RegexEnvFile.FindStringIndex(st) != nil {
		return st
	}
	return "local.env"
}

func init() {
	RegexEnvFile = regexp.MustCompile("^(\\.\\./|[/a-zA-Z])+(/[a-zA-Z\\d.\\-])*\\.env$")
	RegexDatabase = regexp.MustCompile("^[A-Za-z]+[A-Za-z0-9]+$")
}

func Initialize(cfgPath string, pDatabaseName string) {
	if len(uri) > 0 {
		return
	}

	if RegexDatabase.MatchString(pDatabaseName) {
		databaseName = pDatabaseName
	} else {
		log.Panicf("Database >%v< is not a valid database name", pDatabaseName)
	}

	var envFile = resolveFile(cfgPath)
	if err := godotenv.Load(envFile); err != nil {
		log.Println("No .env file found")
		log.Panicf("Impossible to Initialize:%v", err)
	}

	uri = os.Getenv("MONGODB_URI")

	if uri == "" {
		log.Fatal("You must set your 'MONGODB_URI' environmental variable. " +
			"See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}
	fmt.Println("MONGO URI=" + uri) //.Collection("movies")
}

func mkClient() *mongo.Client {
	var err error = nil
	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err == nil {
		return client
	}
	panic(err)
}

func GetClient() *mongo.Client {
	return conv.Nvl(client).OrCall(mkClient)
}

func mkDatabase() *mongo.Database {
	database = GetClient().Database(databaseName)
	if database == nil {
		log.Panicf("No such database: %v", databaseName)
	}
	return database
}

func GetDatabase() *mongo.Database {
	return conv.Nvl(database).OrCall(mkDatabase)
}

func Shutdown() {
	conv.Nvl(client).DoIfPresent(func(cl *mongo.Client) {
		if err := cl.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	})
}
