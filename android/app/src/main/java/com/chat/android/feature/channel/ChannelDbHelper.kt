package com.chat.android.feature.channel

import android.content.ContentValues
import android.content.Context
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper

data class DbChannel(
    val channelID: String,
    val channelName: String,
    val channelDescription: String,
    val myname: String,
    val myimg: String,
    val displayStatus: Int,
    val invitationCode: String,
    val invitationGuestCode: String
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
                myimg TEXT,
                displayStatus INTEGER,
                invitationCode TEXT,
                invitationGuestCode TEXT
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
    }

    fun saveChannel(channel: DbChannel) {
        val values = ContentValues().apply {
            put("channelID", channel.channelID)
            put("channelName", channel.channelName)
            put("channelDescription", channel.channelDescription)
            put("myname", channel.myname)
            put("myimg", channel.myimg)
            put("displayStatus", channel.displayStatus)
            put("invitationCode", channel.invitationCode)
            put("invitationGuestCode", channel.invitationGuestCode)
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
                    myimg = it.getString(it.getColumnIndexOrThrow("myimg")),
                    displayStatus = it.getInt(it.getColumnIndexOrThrow("displayStatus")),
                    invitationCode = it.getString(it.getColumnIndexOrThrow("invitationCode")),
                    invitationGuestCode = it.getString(it.getColumnIndexOrThrow("invitationGuestCode"))
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
                        myimg = it.getString(it.getColumnIndexOrThrow("myimg")),
                        displayStatus = it.getInt(it.getColumnIndexOrThrow("displayStatus")),
                        invitationCode = it.getString(it.getColumnIndexOrThrow("invitationCode")),
                        invitationGuestCode = it.getString(it.getColumnIndexOrThrow("invitationGuestCode"))
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
        private const val DATABASE_VERSION = 2
    }
}
