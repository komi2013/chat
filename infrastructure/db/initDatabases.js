// === MongoDB 初期構築スクリプト ===
// bashで
// mongosh < initDatabases.js
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

// すべてのDBを初期化（空でOK）
dbNames.forEach(name => {
  const targetDB = db.getSiblingDB(name);
  print(`✅ Initialized database: ${name}`);

  // chatSequenceのみ初期データを投入
  if (name === "chatSequence") {
    const sequenceCol = targetDB.getCollection("sequence");

    // 既存データをクリア（再初期化時を考慮）
    sequenceCol.deleteMany({});

    const now = new Date();

    const initialData = [
      {
        _id: "channelID",
        lock: 0,
        count: "0",
        description: "channelIDのシーケンス",
        updatedAt: now,
      },
      {
        _id: "userID",
        lock: 0,
        count: "0",
        description: "userIDのシーケンス",
        updatedAt: now,
      },
      {
        _id: "tweetID",
        lock: 0,
        count: "0",
        description: "tweetIDのシーケンス",
        updatedAt: now,
      },
    ];

    sequenceCol.insertMany(initialData);
    print("🚀 Inserted initial sequence data into chatSequence.sequence");
  }
});
