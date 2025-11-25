// bash
// mongosh "mongodb://root:12345678@mongo:27017/?retryWrites=true&w=majority&authSource=admin" < initDatabases.js

const dbNames = [
  "chatAd",
  "chatAdPrice",
  "chatChannel",
  "chatFile",
  "chatNickname",
  "chatReception",
  "chatSequence",
  "chatSession",
  "chatTweet",
  "chatUser",
];

// convert DB name to collection name
function collectionName(dbName) {
  const trimmed = dbName.replace(/^chat/, "");   // remove chat prefix
  return trimmed.charAt(0).toLowerCase() + trimmed.slice(1); // lowercase first letter
}

dbNames.forEach(name => {
  const targetDB = db.getSiblingDB(name);
  const colName = collectionName(name);

  print(`📦 Init DB: ${name}, Collection: ${colName}`);

  // create empty collection to ensure DB exists
  targetDB.createCollection(colName);

  if (name === "chatSequence") {
    const sequenceCol = targetDB.getCollection(colName);
    sequenceCol.deleteMany({});

    const now = new Date();

    const initialData = [
      { _id: "adID", lock: 0, count: "5", description: "adIDのシーケンス", updatedAt: now },
      { _id: "channelID", lock: 0, count: "A", description: "channelIDのシーケンス", updatedAt: now },
      { _id: "fileID", lock: 0, count: "D", description: "fileIDのシーケンス", updatedAt: now },
      { _id: "tweetID", lock: 0, count: "Y", description: "tweetIDのシーケンス", updatedAt: now },
      { _id: "userID", lock: 0, count: "2", description: "userIDのシーケンス", updatedAt: now },
    ];

    sequenceCol.insertMany(initialData);
    print("🚀 Inserted initial sequence data");
  }
});
