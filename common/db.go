package common

import (
    "context"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type DBs struct {
  AdDB *mongo.Database
  AdPriceDB *mongo.Database
  UserDB    *mongo.Database
  SessionDB *mongo.Database

}

var DB *DBs

func InitMongo() {
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  adClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoAd))
  if err != nil {
    log.Fatalf("Mongo UserClient connect error: %v", err)
  }

  adPriceClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoAdPrice))
  if err != nil {
    log.Fatalf("Mongo UserClient connect error: %v", err)
  }

  userClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoUser))
  if err != nil {
    log.Fatalf("Mongo UserClient connect error: %v", err)
  }

  sessionClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoSession))
  if err != nil {
    log.Fatalf("Mongo UserClient connect error: %v", err)
  }

  DB = &DBs{
    AdDB: adClient.Database("chatAd"),
    AdPriceDB: adPriceClient.Database("chatAdPrice"),
    UserDB:    userClient.Database("chatUser"),
    SessionDB: sessionClient.Database("chatSession"),

  }
}
