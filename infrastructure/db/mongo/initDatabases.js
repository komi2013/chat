// === MongoDB 初期構築スクリプト ===
// bashで
// mongosh \
//   -u root \
//   -p 12345678 \
//   --authenticationDatabase admin \
//   < initDatabases.js

const dbNames = [
  "chatAd",
  "chatAdPrice",
  "chatChannel",
  "chatFile",
  "chatInvoice",
  "chatNickname",
  "chatReception",
  "chatSequence",
  "chatSession",
  "chatTweet",
  "chatUser",
];

// Current timestamp for all initial data
const now = new Date();

// Map DB name → collection name (without "chat" prefix, lowercase first letter)
const collectionMap = {
  chatAd: "ad",
  chatAdPrice: "adPrice",
  chatChannel: "channel",
  chatFile: "file",
  chatInvoice: "invoice",
  chatNickname: "nickname",
  chatReception: "reception",
  chatSequence: "sequence",
  chatSession: "session",
  chatTweet: "tweet",
  chatUser: "user",
};

dbNames.forEach(name => {
  const targetDB = db.getSiblingDB(name);
  const colName = collectionMap[name];

  print(`✅ Initializing database: ${name}, collection: ${colName}`);

  // Special initialization for chatSequence
  if (name === "chatSequence") {
    const sequenceCol = targetDB.getCollection(colName);
    sequenceCol.deleteMany({});
    const sequenceData = [
      { _id: "adID", lock: 0, count: "5", description: "adIDのシーケンス", updatedAt: now },
      { _id: "channelID", lock: 0, count: "A", description: "channelIDのシーケンス", updatedAt: now },
      { _id: "invoiceID", lock: 0, count: "0", description: "invoiceIDのシーケンス", updatedAt: now },
      { _id: "tweetID", lock: 0, count: "Y", description: "tweetIDのシーケンス", updatedAt: now },
      { _id: "userID", lock: 0, count: "2", description: "userIDのシーケンス", updatedAt: now },
    ];
    sequenceCol.insertMany(sequenceData);
    print("🚀 Inserted initial sequence data into chatSequence.sequence");
  }

  // Special initialization for chatAdPrice
  else if (name === "chatAdPrice") {
    const adPriceCol = targetDB.getCollection(colName);
    adPriceCol.deleteMany({});
    const adPriceData = [
      {
        _id: '0',
        latitudeNorth: 0,
        latitudeSouth: 0,
        longitudeEast: 0,
        longitudeWest: 0,
        adStart: 100,
        adEnd: 724,
        adPriceYen: 800,
        updatedAt: now,
      },
    ];
    adPriceCol.insertMany(adPriceData);
    print("🚀 Inserted initial data into chatAdPrice.adPrice");
  }

  // Other DBs: create empty collection
  else {
    const col = targetDB.getCollection(colName);
    col.insertOne({ initializedAt: now });
    col.deleteMany({});
    print(`🟢 Created empty collection ${colName} in database ${name}`);
  }
});

print("🎉 All databases and collections initialized successfully!");
