package com.chat.android.feature.channel

import android.content.ContentValues
import android.content.Context
import android.database.Cursor
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper

/**
 * channel テーブルの1行。
 *
 * 招待コード（invitationCode / invitationGuestCode）は保持しない。
 * サーバーは ChannelEdit/ の generateInvitation を受けるたびにコードを
 * 回転させる（controller/ChannelEdit.go:134）ため、保存すると必ず古くなる。
 * 必要になったとき API から取得して画面上でだけ扱う。
 * displayStatus も現時点で未使用のため持ちない。
 */
data class DbChannel(
    val channelID: String,
    val channelName: String,
    val channelDescription: String,
    val myname: String,
    val myimg: String
)

data class DbAlias(
    val aliasID: String,
    val channelID: String,
    val aliasName: String,
    val aliasImg: String,
    val userID: String,
    val aliasBio: String,
    val accessRight: String
)

data class DbGroup(
    val groupID: String,
    val channelID: String,
    val groupName: String,
    val groupImg: String,
    val aliasNamesJson: String, // JSON array string e.g. '["name1", "name2"]'
    val groupBio: String
)

data class DbThread(
    val messageID: String,
    val parentID: String,
    val channelID: String,
    val messageTxt: String,
    val aliasName: String,
    val aliasImg: String,
    val aliasNamesJson: String,
    val backID: String,
    val emojisJson: String,
    val createdAt: String
)

data class DbThreadHead(
    val parentID: String,
    val channelID: String,
    val title: String,
    val messageTxt: String,
    val description: String,
    val aliasName: String,
    val aliasNamesJson: String,
    val adminNamesJson: String,
    val displayStatus: Int,
    val broadcastFlag: Int,
    val updatedAt: String,
    /** リアクション。vue の threadHead.emojis に対応する。 */
    val emojisJson: String = "[]",
    /** 返信元のスレッドID。vue の Channel.vue は !backID で絞り込み、
     *  一覧にはトップレベルのスレッドだけを出す。 */
    val backID: String = ""
)

class ChannelDbHelper(context: Context) : SQLiteOpenHelper(context, DATABASE_NAME, null, DATABASE_VERSION) {

    override fun onCreate(db: SQLiteDatabase) {
        db.execSQL("""
            CREATE TABLE channel (
                channelID TEXT PRIMARY KEY,
                channelName TEXT,
                channelDescription TEXT,
                myname TEXT,
                myimg TEXT
            )
        """.trimIndent())

        db.execSQL("""
            CREATE TABLE alias (
                aliasID TEXT PRIMARY KEY,
                channelID TEXT,
                aliasName TEXT,
                aliasImg TEXT,
                userID TEXT,
                aliasBio TEXT,
                accessRight TEXT
            )
        """.trimIndent())

        db.execSQL("""
            CREATE TABLE "group" (
                groupID TEXT PRIMARY KEY,
                channelID TEXT,
                groupName TEXT,
                groupImg TEXT,
                aliasNames TEXT,
                groupBio TEXT
            )
        """.trimIndent())

        db.execSQL("""
            CREATE TABLE user_profile (
                profileId TEXT NOT NULL PRIMARY KEY,
                name TEXT,
                mail TEXT,
                nickname TEXT,
                channelID TEXT,
                accessRight TEXT,
                admin INTEGER,
                telephone TEXT,
                walletAddress TEXT,
                latitude REAL,
                longitude REAL,
                nickImg TEXT,
                nickBio TEXT
            )
        """.trimIndent())

        db.execSQL("""
            CREATE TABLE user_nickname (
                nickname TEXT NOT NULL PRIMARY KEY,
                nickImg TEXT,
                nickBio TEXT,
                good INTEGER NOT NULL DEFAULT 0,
                bad INTEGER NOT NULL DEFAULT 0,
                createdAt TEXT,
                accessRight TEXT
            )
        """.trimIndent())

        db.execSQL(THREAD_HEAD_TABLE)
        db.execSQL(THREAD_TABLE)
    }

    override fun onUpgrade(db: SQLiteDatabase, oldVersion: Int, newVersion: Int) {
        // v2 で user_profile / user_nickname を追加。
        // channel / alias / "group" は既存データを保持する。
        if (oldVersion < 2) {
            db.execSQL(
                """
                CREATE TABLE IF NOT EXISTS user_profile (
                    profileId TEXT NOT NULL PRIMARY KEY,
                    name TEXT,
                    mail TEXT,
                    nickname TEXT,
                    channelID TEXT,
                    accessRight TEXT,
                    admin INTEGER,
                    telephone TEXT,
                    walletAddress TEXT,
                    latitude REAL,
                    longitude REAL,
                    nickImg TEXT,
                    nickBio TEXT
                )
                """.trimIndent()
            )
            db.execSQL(
                """
                CREATE TABLE IF NOT EXISTS user_nickname (
                    nickname TEXT NOT NULL PRIMARY KEY,
                    nickImg TEXT,
                    nickBio TEXT,
                    good INTEGER NOT NULL DEFAULT 0,
                    bad INTEGER NOT NULL DEFAULT 0,
                    createdAt TEXT,
                    accessRight TEXT
                )
                """.trimIndent()
            )
        }

        // v3 で channel から displayStatus / invitationCode / invitationGuestCode を削除。
        // ALTER TABLE ... DROP COLUMN は SQLite 3.35+（Android 13 / API 33 以降）が必要なため、
        // minSdk 24 では使えない。テーブルを作り直して既存行を移す方式にする。
        if (oldVersion < 3) {
            db.execSQL(
                """
                CREATE TABLE channel_new (
                    channelID TEXT PRIMARY KEY,
                    channelName TEXT,
                    channelDescription TEXT,
                    myname TEXT,
                    myimg TEXT
                )
                """.trimIndent()
            )
            db.execSQL(
                """
                INSERT INTO channel_new (channelID, channelName, channelDescription, myname, myimg)
                SELECT channelID, channelName, channelDescription, myname, myimg FROM channel
                """.trimIndent()
            )
            db.execSQL("DROP TABLE channel")
            db.execSQL("ALTER TABLE channel_new RENAME TO channel")
        }

        // v4 でスレッド用の threadHead / thread を追加（vue の IndexedDB ストアに対応）。
        if (oldVersion < 4) {
            db.execSQL(THREAD_HEAD_TABLE)
            db.execSQL(THREAD_TABLE)
        }

        // v5 で threadHead に emojis（リアクション）を追加。
        // 注意: v4 の CREATE TABLE にも emojis 列を含めているため、
        // v3 から直接バージョンアップするとこの列は既に存在している。
        // 無条件に ADD COLUMN すると duplicate column で例外になり、
        // onUpgrade 全体がロールバックしてDBが開かなくなるため、
        // 存在を確認してから追加する。
        if (oldVersion < 5) {
            val hasEmojis = db.rawQuery("PRAGMA table_info(threadHead)", null).use { cursor ->
                (0 until cursor.count).any {
                    cursor.getString(it).equals("emojis", ignoreCase = true)
                }
            }
            if (!hasEmojis) {
                db.execSQL("ALTER TABLE threadHead ADD COLUMN emojis TEXT")
            }
        }

        // v6 で threadHead に backID を追加。
        if (oldVersion < 6) {
            db.execSQL("ALTER TABLE threadHead ADD COLUMN backID TEXT")
        }
    }

    /** スレッドの親（見出し）を保存。vue の threadHead IndexedDB ストアに対応。 */
    fun saveThreadHead(head: DbThreadHead) {
        val values = ContentValues().apply {
            put("parentID", head.parentID)
            put("channelID", head.channelID)
            put("title", head.title)
            put("messageTxt", head.messageTxt)
            put("description", head.description)
            put("aliasName", head.aliasName)
            put("aliasNames", head.aliasNamesJson)
            put("adminNames", head.adminNamesJson)
            put("displayStatus", head.displayStatus)
            put("broadcastFlag", head.broadcastFlag)
            put("updatedAt", head.updatedAt)
            put("emojis", head.emojisJson)
            put("backID", head.backID)
        }
        writableDatabase.insertWithOnConflict("threadHead", null, values, SQLiteDatabase.CONFLICT_REPLACE)
    }

    fun getThreadHead(parentID: String): DbThreadHead? {
        val cursor = readableDatabase.query(
            "threadHead", null, "parentID = ?", arrayOf(parentID), null, null, null
        )
        return cursor.use {
            if (it.moveToFirst()) it.toThreadHead() else null
        }
    }

    fun saveThread(thread: DbThread) {
        val values = ContentValues().apply {
            put("messageID", thread.messageID)
            put("parentID", thread.parentID)
            put("channelID", thread.channelID)
            put("messageTxt", thread.messageTxt)
            put("aliasName", thread.aliasName)
            put("aliasImg", thread.aliasImg)
            put("aliasNames", thread.aliasNamesJson)
            put("backID", thread.backID)
            put("emojis", thread.emojisJson)
            put("createdAt", thread.createdAt)
        }
        writableDatabase.insertWithOnConflict("thread", null, values, SQLiteDatabase.CONFLICT_REPLACE)
    }

    fun getThread(messageID: String): DbThread? {
        val cursor = readableDatabase.query(
            "thread", null, "messageID = ?", arrayOf(messageID), null, null, null
        )
        return cursor.use {
            if (it.moveToFirst()) it.toThread() else null
        }
    }

    fun deleteThread(messageID: String) {
        writableDatabase.delete("thread", "messageID = ?", arrayOf(messageID))
    }

    /** 親ID（スレッドID）に属するメッセージを古い順に返す。 */
    fun getThreads(parentID: String): List<DbThread> {
        val list = mutableListOf<DbThread>()
        val cursor = readableDatabase.query(
            "thread", null, "parentID = ?", arrayOf(parentID), null, null, "createdAt ASC, messageID ASC"
        )
        cursor.use {
            while (it.moveToNext()) list.add(it.toThread())
        }
        return list
    }

    /**
 * チャネルのスレッド一覧。vue の Channel.vue と同じく
 * backID が無いトップレベルのスレッドだけを、更新日時順に返す。
 */
    fun getThreadHeadsForChannel(channelID: String): List<DbThreadHead> {
        val list = mutableListOf<DbThreadHead>()
        val cursor = readableDatabase.query(
            "threadHead",
            null,
            "channelID = ? AND (backID IS NULL OR backID = '')",
            arrayOf(channelID),
            null,
            null,
            "updatedAt DESC"
        )
        cursor.use {
            while (it.moveToNext()) list.add(it.toThreadHead())
        }
        return list
    }

    private fun Cursor.toThreadHead() = DbThreadHead(
        parentID = getString(getColumnIndexOrThrow("parentID")),
        channelID = getString(getColumnIndexOrThrow("channelID")),
        title = getString(getColumnIndexOrThrow("title")),
        messageTxt = getString(getColumnIndexOrThrow("messageTxt")),
        description = getString(getColumnIndexOrThrow("description")),
        aliasName = getString(getColumnIndexOrThrow("aliasName")),
        aliasNamesJson = getString(getColumnIndexOrThrow("aliasNames")),
        adminNamesJson = getString(getColumnIndexOrThrow("adminNames")),
        displayStatus = getInt(getColumnIndexOrThrow("displayStatus")),
        broadcastFlag = getInt(getColumnIndexOrThrow("broadcastFlag")),
        updatedAt = getString(getColumnIndexOrThrow("updatedAt")),
        emojisJson = getString(getColumnIndexOrThrow("emojis")),
        backID = getString(getColumnIndexOrThrow("backID"))
    )

    private fun Cursor.toThread() = DbThread(
        messageID = getString(getColumnIndexOrThrow("messageID")),
        parentID = getString(getColumnIndexOrThrow("parentID")),
        channelID = getString(getColumnIndexOrThrow("channelID")),
        messageTxt = getString(getColumnIndexOrThrow("messageTxt")),
        aliasName = getString(getColumnIndexOrThrow("aliasName")),
        aliasImg = getString(getColumnIndexOrThrow("aliasImg")),
        aliasNamesJson = getString(getColumnIndexOrThrow("aliasNames")),
        backID = getString(getColumnIndexOrThrow("backID")),
        emojisJson = getString(getColumnIndexOrThrow("emojis")),
        createdAt = getString(getColumnIndexOrThrow("createdAt"))
    )

    fun saveChannel(channel: DbChannel) {
        val values = ContentValues().apply {
            put("channelID", channel.channelID)
            put("channelName", channel.channelName)
            put("channelDescription", channel.channelDescription)
            put("myname", channel.myname)
            put("myimg", channel.myimg)
        }
        writableDatabase.insertWithOnConflict("channel", null, values, SQLiteDatabase.CONFLICT_REPLACE)
    }

    fun getChannel(channelID: String): DbChannel? {
        val cursor = readableDatabase.query(
            "channel", null, "channelID = ?", arrayOf(channelID), null, null, null
        )
        return cursor.use {
            if (it.moveToFirst()) {
                DbChannel(
                    channelID = it.getString(it.getColumnIndexOrThrow("channelID")),
                    channelName = it.getString(it.getColumnIndexOrThrow("channelName")),
                    channelDescription = it.getString(it.getColumnIndexOrThrow("channelDescription")),
                    myname = it.getString(it.getColumnIndexOrThrow("myname")),
                    myimg = it.getString(it.getColumnIndexOrThrow("myimg"))
                )
            } else null
        }
    }

    fun getAllChannels(): List<DbChannel> {
        val list = mutableListOf<DbChannel>()
        val cursor = readableDatabase.query("channel", null, null, null, null, null, null)
        cursor.use {
            while (it.moveToNext()) {
                list.add(
                    DbChannel(
                        channelID = it.getString(it.getColumnIndexOrThrow("channelID")),
                        channelName = it.getString(it.getColumnIndexOrThrow("channelName")),
                        channelDescription = it.getString(it.getColumnIndexOrThrow("channelDescription")),
                        myname = it.getString(it.getColumnIndexOrThrow("myname")),
                        myimg = it.getString(it.getColumnIndexOrThrow("myimg"))
                    )
                )
            }
        }
        return list
    }

    fun deleteChannel(channelID: String) {
        writableDatabase.delete("channel", "channelID = ?", arrayOf(channelID))
    }

    fun saveAlias(alias: DbAlias) {
        val values = ContentValues().apply {
            put("aliasID", alias.aliasID)
            put("channelID", alias.channelID)
            put("aliasName", alias.aliasName)
            put("aliasImg", alias.aliasImg)
            put("userID", alias.userID)
            put("aliasBio", alias.aliasBio)
            put("accessRight", alias.accessRight)
        }
        writableDatabase.insertWithOnConflict("alias", null, values, SQLiteDatabase.CONFLICT_REPLACE)
    }

    fun getAliasesForChannel(channelID: String): List<DbAlias> {
        val list = mutableListOf<DbAlias>()
        val cursor = readableDatabase.query(
            "alias", null, "channelID = ?", arrayOf(channelID), null, null, null
        )
        cursor.use {
            while (it.moveToNext()) {
                list.add(
                    DbAlias(
                        aliasID = it.getString(it.getColumnIndexOrThrow("aliasID")),
                        channelID = it.getString(it.getColumnIndexOrThrow("channelID")),
                        aliasName = it.getString(it.getColumnIndexOrThrow("aliasName")),
                        aliasImg = it.getString(it.getColumnIndexOrThrow("aliasImg")),
                        userID = it.getString(it.getColumnIndexOrThrow("userID")),
                        aliasBio = it.getString(it.getColumnIndexOrThrow("aliasBio")),
                        accessRight = it.getString(it.getColumnIndexOrThrow("accessRight"))
                    )
                )
            }
        }
        return list
    }

    fun deleteAlias(aliasID: String) {
        writableDatabase.delete("alias", "aliasID = ?", arrayOf(aliasID))
    }

    fun saveGroup(group: DbGroup) {
        val values = ContentValues().apply {
            put("groupID", group.groupID)
            put("channelID", group.channelID)
            put("groupName", group.groupName)
            put("groupImg", group.groupImg)
            put("aliasNames", group.aliasNamesJson)
            put("groupBio", group.groupBio)
        }
        writableDatabase.insertWithOnConflict("\"group\"", null, values, SQLiteDatabase.CONFLICT_REPLACE)
    }

    /** 指定 groupID のローカル保存済み groupBio を返す。未保存なら null。 */
    fun getGroupBio(groupID: String): String? {
        readableDatabase.query(
            "\"group\"", arrayOf("groupBio"), "groupID = ?", arrayOf(groupID), null, null, null
        ).use {
            return if (it.moveToFirst()) it.getString(0) else null
        }
    }

    fun getGroupsForChannel(channelID: String): List<DbGroup> {
        val list = mutableListOf<DbGroup>()
        val cursor = readableDatabase.query(
            "\"group\"", null, "channelID = ?", arrayOf(channelID), null, null, null
        )
        cursor.use {
            while (it.moveToNext()) {
                list.add(
                    DbGroup(
                        groupID = it.getString(it.getColumnIndexOrThrow("groupID")),
                        channelID = it.getString(it.getColumnIndexOrThrow("channelID")),
                        groupName = it.getString(it.getColumnIndexOrThrow("groupName")),
                        groupImg = it.getString(it.getColumnIndexOrThrow("groupImg")),
                        aliasNamesJson = it.getString(it.getColumnIndexOrThrow("aliasNames")),
                        groupBio = it.getString(it.getColumnIndexOrThrow("groupBio"))
                    )
                )
            }
        }
        return list
    }

    fun deleteGroup(groupID: String) {
        writableDatabase.delete("\"group\"", "groupID = ?", arrayOf(groupID))
    }

    companion object {
        private const val DATABASE_NAME = "ChannelFeatureDB.db"
        private const val DATABASE_VERSION = 6

/** スレッド見出し。vue の threadHead IndexedDB ストアの項目に揃えている。 */
private val THREAD_HEAD_TABLE = """
        CREATE TABLE IF NOT EXISTS threadHead (
            parentID TEXT NOT NULL PRIMARY KEY,
            channelID TEXT,
            title TEXT,
            messageTxt TEXT,
            description TEXT,
            aliasName TEXT,
            aliasNames TEXT,
            adminNames TEXT,
            displayStatus INTEGER NOT NULL DEFAULT 0,
            broadcastFlag INTEGER NOT NULL DEFAULT 0,
            updatedAt TEXT,
            emojis TEXT,
            backID TEXT
        )
    """.trimIndent()

/** スレッドのメッセージ。vue の thread IndexedDB ストアの項目に揃えている。 */
private val THREAD_TABLE = """
        CREATE TABLE IF NOT EXISTS thread (
            messageID TEXT NOT NULL PRIMARY KEY,
            parentID TEXT,
            channelID TEXT,
            messageTxt TEXT,
            aliasName TEXT,
            aliasImg TEXT,
            aliasNames TEXT,
            backID TEXT,
            emojis TEXT,
            createdAt TEXT
        )
    """.trimIndent()
    }
}
