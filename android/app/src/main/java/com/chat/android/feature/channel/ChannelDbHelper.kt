package com.chat.android.feature.channel

import android.content.ContentValues
import android.content.Context
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
    }

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
        private const val DATABASE_VERSION = 3
    }
}
