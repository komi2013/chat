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
  ChannelDB *mongo.Database
  FileDB *mongo.Database
  InvoiceDB *mongo.Database
  NicknameDB *mongo.Database
  ReceptionDB *mongo.Database
  SequenceDB *mongo.Database
  SessionDB *mongo.Database
  TweetDB *mongo.Database
  UserDB    *mongo.Database

}

var DB *DBs

func InitMongo() {
  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  adClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoAd))
  if err != nil {
    log.Fatalf("Mongo adClient connect error: %v", err)
  }

  adPriceClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoAdPrice))
  if err != nil {
    log.Fatalf("Mongo adPriceClient connect error: %v", err)
  }

  channelClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoChannel))
  if err != nil {
    log.Fatalf("Mongo channelClient connect error: %v", err)
  }

  fileClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoFile))
  if err != nil {
    log.Fatalf("Mongo fileClient connect error: %v", err)
  }

  invoiceClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoInvoice))
  if err != nil {
    log.Fatalf("Mongo invoiceClient connect error: %v", err)
  }

  nicknameClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoNickname))
  if err != nil {
    log.Fatalf("Mongo nicknameClient connect error: %v", err)
  }

  receptionClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoReception))
  if err != nil {
    log.Fatalf("Mongo receptionClient connect error: %v", err)
  }

  sequenceClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoSequence))
  if err != nil {
    log.Fatalf("Mongo sequenceClient connect error: %v", err)
  }

  sessionClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoSession))
  if err != nil {
    log.Fatalf("Mongo sessionClient connect error: %v", err)
  }

  tweetClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoTweet))
  if err != nil {
    log.Fatalf("Mongo sessionClient connect error: %v", err)
  }

  userClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoUser))
  if err != nil {
    log.Fatalf("Mongo userClient connect error: %v", err)
  }

  DB = &DBs{
    AdDB: adClient.Database("chatAd"),
    AdPriceDB: adPriceClient.Database("chatAdPrice"),
    ChannelDB: channelClient.Database("chatChannel"),
    FileDB: fileClient.Database("chatFile"),
    InvoiceDB: invoiceClient.Database("chatInvoice"),
    NicknameDB: nicknameClient.Database("chatNickname"),
    ReceptionDB: receptionClient.Database("chatReception"),
    SequenceDB: sequenceClient.Database("chatSequence"),
    SessionDB: sessionClient.Database("chatSession"),
    TweetDB: tweetClient.Database("chatTweet"),
    UserDB:    userClient.Database("chatUser"),

  }
}
