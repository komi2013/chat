// === MongoDB 初期構築スクリプト ===
// bashで
// mongosh < initDatabases.js
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

// すべてのDBを初期化（空でOK）
dbNames.forEach(name => {
  const targetDB = db.getSiblingDB(name);
  print(`✅ Initialized database: ${name}`);
  const now = new Date();
  let initialData;
  if (name === "chatSequence") {
    const sequenceCol = targetDB.getCollection("sequence");
    sequenceCol.deleteMany({});
    initialData = [
      {
        _id: "adID",
        lock: 0,
        count: "5",
        description: "adIDのシーケンス",
        updatedAt: now,
      },
      {
        _id: "channelID",
        lock: 0,
        count: "A",
        description: "channelIDのシーケンス",
        updatedAt: now,
      },
      {
        _id: "invoiceID",
        lock: 0,
        count: "0",
        description: "invoiceIDのシーケンス",
        updatedAt: now,
      },
      {
        _id: "tweetID",
        lock: 0,
        count: "Y",
        description: "tweetIDのシーケンス",
        updatedAt: now,
      },
      {
        _id: "userID",
        lock: 0,
        count: "2",
        description: "userIDのシーケンス",
        updatedAt: now,
      },
    ];
    sequenceCol.insertMany(initialData);
    print("🚀 Inserted initial sequence data into chatSequence.sequence");
  }
  if (name === "chatAdPrice") {
    const adPriceCol = targetDB.getCollection("adPrice");
    adPriceCol.deleteMany({});
    initialData = [
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
    ]
  }
});
