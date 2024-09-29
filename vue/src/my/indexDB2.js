// IndexedDBに接続
function openDatabase() {
    return new Promise((resolve, reject) => {
        const request = indexedDB.open('YourDatabaseName', 1);

        request.onupgradeneeded = event => {
            const db = event.target.result;
            const store = db.createObjectStore('timestamp', { keyPath: 'id', autoIncrement: true });

            // 複合インデックスを作成
            store.createIndex('channelID_aliasName', ['channelID', 'aliasName'], { unique: false });
        };

        request.onsuccess = event => resolve(event.target.result);
        request.onerror = event => reject(event.target.error);
    });
}

// データを取得
function getRecords(db, channelID, aliasName) {
    return new Promise((resolve, reject) => {
        const transaction = db.transaction('timestamp', 'readonly');
        const store = transaction.objectStore('timestamp');
        const index = store.index('channelID_aliasName');

        // channelIDとaliasNameの組み合わせで検索
        const request = index.getAll([channelID, aliasName]);

        request.onsuccess = event => resolve(event.target.result);
        request.onerror = event => reject(event.target.error);
    });
}

// データベースを開き、クエリを実行する
openDatabase().then(db => {
    getRecords(db, 'yourChannelID', 'yourAliasName').then(records => {
        console.log('Matching records:', records);
    }).catch(error => {
        console.error('Error retrieving records:', error);
    });
}).catch(error => {
    console.error('Error opening database:', error);
});
