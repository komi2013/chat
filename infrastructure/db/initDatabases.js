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
        _id: "fileID",
        lock: 0,
        count: "D",
        description: "fileIDのシーケンス",
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
});
