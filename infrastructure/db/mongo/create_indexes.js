// MongoDB インデックス作成スクリプト (JavaScript)
// 使用方法: mongo <connection_string> create_indexes.js
// または: mongosh <connection_string> create_indexes.js

// ==========================================
// 高優先度インデックス
// ==========================================
print("=== 高優先度インデックス作成開始 (複合インデックスを単一に分割) ===");

// 1. session.userID
try {
  db = db.getSiblingDB("chatSession");
  db.session.createIndex({ "userID": 1 }, { name: "userID_1" });
  print("✓ 作成成功: session.userID");
} catch (e) {
  if (e.message.includes("already exists") || e.message.includes("duplicate")) {
    print("✓ スキップ: session.userID (既に存在)");
  } else {
    print("✗ 失敗: session.userID - " + e.message);
  }
}

// 2. nickname.userID
try {
  db = db.getSiblingDB("chatNickname");
  db.nickname.createIndex({ "userID": 1 }, { name: "userID_1" });
  print("✓ 作成成功: nickname.userID");
} catch (e) {
  if (e.message.includes("already exists") || e.message.includes("duplicate")) {
    print("✓ スキップ: nickname.userID (既に存在)");
  } else {
    print("✗ 失敗: nickname.userID - " + e.message);
  }
}

// 3. ad.userID
try {
  db = db.getSiblingDB("chatAd");
  db.ad.createIndex({ "userID": 1 }, { name: "userID_1" });
  print("✓ 作成成功: ad.userID");
} catch (e) {
  if (e.message.includes("already exists") || e.message.includes("duplicate")) {
    print("✓ スキップ: ad.userID (既に存在)");
  } else {
    print("✗ 失敗: ad.userID - " + e.message);
  }
}

// 4. ad.adYen (降順)
try {
  db = db.getSiblingDB("chatAd");
  db.ad.createIndex({ "adYen": -1 }, { name: "adYen_-1" });
  print("✓ 作成成功: ad.adYen (降順)");
} catch (e) {
  if (e.message.includes("already exists") || e.message.includes("duplicate")) {
    print("✓ スキップ: ad.adYen (既に存在)");
  } else {
    print("✗ 失敗: ad.adYen - " + e.message);
  }
}

// 5. file.updatedAt (複合インデックスから分割)
try {
  db = db.getSiblingDB("chatFile");
  db.file.createIndex({ "updatedAt": 1 }, { name: "updatedAt_1" });
  print("✓ 作成成功: file.updatedAt (単一)");
} catch (e) {
  if (e.message.includes("already exists") || e.message.includes("duplicate")) {
    print("✓ スキップ: file.updatedAt (既に存在)");
  } else {
    print("✗ 失敗: file.updatedAt - " + e.message);
  }
}

// 6. file.usageType (複合インデックスから分割)
try {
  db = db.getSiblingDB("chatFile");
  db.file.createIndex({ "usageType": 1 }, { name: "usageType_1" });
  print("✓ 作成成功: file.usageType (単一)");
} catch (e) {
  if (e.message.includes("already exists") || e.message.includes("duplicate")) {
    print("✓ スキップ: file.usageType (既に存在)");
  } else {
    print("✗ 失敗: file.usageType - " + e.message);
  }
}

// 7. tweet.updatedAt (降順)
try {
  db = db.getSiblingDB("chatTweet");
  db.tweet.createIndex({ "updatedAt": -1 }, { name: "updatedAt_-1" });
  print("✓ 作成成功: tweet.updatedAt (降順)");
} catch (e) {
  if (e.message.includes("already exists") || e.message.includes("duplicate")) {
    print("✓ スキップ: tweet.updatedAt (既に存在)");
  } else {
    print("✗ 失敗: tweet.updatedAt - " + e.message);
  }
}

// 8. ad.distance (複合インデックスから分割)
try {
  db = db.getSiblingDB("chatAd");
  db.ad.createIndex({ "distance": 1 }, { name: "distance_1" });
  print("✓ 作成成功: ad.distance (単一)");
} catch (e) {
  if (e.message.includes("already exists") || e.message.includes("duplicate")) {
    print("✓ スキップ: ad.distance (既に存在)");
  } else {
    print("✗ 失敗: ad.distance - " + e.message);
  }
}

// 9. ad.adYen (複合インデックスから分割 - 既に「adYen_-1」が存在するため、これは不要だが、ここでは「adYen」として昇順で追加)
// ただし、adYenのインデックスは既に4番で降順で作成されているため、重複を避けるためにスキップします。

// ==========================================
// 中優先度インデックス
// ==========================================
print("\n=== 中優先度インデックス作成開始 ===");

// 10. user.googleJWTSub
try {
  db = db.getSiblingDB("chatUser");
  db.user.createIndex({ "googleJWTSub": 1 }, { name: "googleJWTSub_1" });
  print("✓ 作成成功: user.googleJWTSub");
} catch (e) {
  if (e.message.includes("already exists") || e.message.includes("duplicate")) {
    print("✓ スキップ: user.googleJWTSub (既に存在)");
  } else {
    print("✗ 失敗: user.googleJWTSub - " + e.message);
  }
}

// 11. invoice.userID
try {
  db = db.getSiblingDB("chatInvoice");
  db.invoice.createIndex({ "userID": 1 }, { name: "userID_1" });
  print("✓ 作成成功: invoice.userID");
} catch (e) {
  if (e.message.includes("already exists") || e.message.includes("duplicate")) {
    print("✓ スキップ: invoice.userID (既に存在)");
  } else {
    print("✗ 失敗: invoice.userID - " + e.message);
  }
}

// 12. invoice.adID
try {
  db = db.getSiblingDB("chatInvoice");
  db.invoice.createIndex({ "adID": 1 }, { name: "adID_1" });
  print("✓ 作成成功: invoice.adID");
} catch (e) {
  if (e.message.includes("already exists") || e.message.includes("duplicate")) {
    print("✓ スキップ: invoice.adID (既に存在)");
  } else {
    print("✗ 失敗: invoice.adID - " + e.message);
  }
}

// 13. invoice.invoiceStatus
try {
  db = db.getSiblingDB("chatInvoice");
  db.invoice.createIndex({ "invoiceStatus": 1 }, { name: "invoiceStatus_1" });
  print("✓ 作成成功: invoice.invoiceStatus");
} catch (e) {
  if (e.message.includes("already exists") || e.message.includes("duplicate")) {
    print("✓ スキップ: invoice.invoiceStatus (既に存在)");
  } else {
    print("✗ 失敗: invoice.invoiceStatus - " + e.message);
  }
}

print("\n=== インデックス作成完了 ===");