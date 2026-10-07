package com.chat.android.feature.settings

import android.content.Context
import android.database.sqlite.SQLiteDatabase
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.Card
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.OutlinedButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONObject

private data class TableData(val name: String, val rows: List<String>)
private data class DatabaseData(val name: String, val tables: List<TableData>)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingsScreen() {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    var databases by remember { mutableStateOf<List<DatabaseData>>(emptyList()) }
    var selectedDatabase by remember { mutableStateOf<DatabaseData?>(null) }
    var selectedTable by remember { mutableStateOf<TableData?>(null) }
    var databaseMenu by remember { mutableStateOf(false) }
    var tableMenu by remember { mutableStateOf(false) }
    var loading by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }

    fun refresh() {
        scope.launch {
            loading = true
            error = null
            try {
                databases = withContext(Dispatchers.IO) { readDatabases(context) }
                selectedDatabase = databases.firstOrNull { it.name == selectedDatabase?.name } ?: databases.firstOrNull()
                selectedTable = selectedDatabase?.tables?.firstOrNull { it.name == selectedTable?.name }
                    ?: selectedDatabase?.tables?.firstOrNull()
            } catch (e: Exception) {
                error = e.message ?: "SQLite データを読み込めませんでした。"
            } finally {
                loading = false
            }
        }
    }

    LaunchedEffect(Unit) { refresh() }

    Column(Modifier.fillMaxSize().padding(16.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
            Column(Modifier.weight(1f)) {
                Text("設定", style = MaterialTheme.typography.headlineSmall)
                Text("この端末に保存された SQLite データを表示します。", style = MaterialTheme.typography.bodySmall)
            }
            TextButton(onClick = { refresh() }, enabled = !loading) { Text("再読み込み") }
        }

        ExposedDropdownMenuBox(expanded = databaseMenu, onExpandedChange = { databaseMenu = it }) {
            OutlinedButton(onClick = { databaseMenu = true }, modifier = Modifier.menuAnchor().fillMaxWidth()) {
                Text(selectedDatabase?.name ?: "データベースを選択", modifier = Modifier.weight(1f))
                ExposedDropdownMenuDefaults.TrailingIcon(expanded = databaseMenu)
            }
            ExposedDropdownMenu(expanded = databaseMenu, onDismissRequest = { databaseMenu = false }) {
                databases.forEach { db ->
                    DropdownMenuItem(text = { Text(db.name) }, onClick = {
                        selectedDatabase = db
                        selectedTable = db.tables.firstOrNull()
                        databaseMenu = false
                    })
                }
            }
        }

        val tables = selectedDatabase?.tables.orEmpty()
        ExposedDropdownMenuBox(expanded = tableMenu, onExpandedChange = { tableMenu = it }) {
            OutlinedButton(onClick = { tableMenu = true }, modifier = Modifier.menuAnchor().fillMaxWidth(), enabled = tables.isNotEmpty()) {
                Text(selectedTable?.let { "${it.name} (${it.rows.size} rows)" } ?: "テーブルなし", modifier = Modifier.weight(1f))
                ExposedDropdownMenuDefaults.TrailingIcon(expanded = tableMenu)
            }
            ExposedDropdownMenu(expanded = tableMenu, onDismissRequest = { tableMenu = false }) {
                tables.forEach { table ->
                    DropdownMenuItem(text = { Text("${table.name} (${table.rows.size})") }, onClick = {
                        selectedTable = table
                        tableMenu = false
                    })
                }
            }
        }

        when {
            loading -> CircularProgressIndicator()
            error != null -> Text(error.orEmpty(), color = MaterialTheme.colorScheme.error)
            databases.isEmpty() -> Text("SQLite データベースはまだありません。")
            tables.isEmpty() -> Text("このデータベースにテーブルはありません。")
            else -> Text("${selectedTable?.rows?.size ?: 0} 件", style = MaterialTheme.typography.labelLarge)
        }

        LazyColumn(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            items(selectedTable?.rows.orEmpty().withIndex().toList(), key = { it.index }) { (index, json) ->
                Card(Modifier.fillMaxWidth()) {
                    Column(Modifier.padding(12.dp)) {
                        Text("#${index + 1}", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.primary)
                        Text(json, fontFamily = FontFamily.Monospace, style = MaterialTheme.typography.bodySmall)
                    }
                }
            }
        }
    }
}


/** Read existing app database files without creating or modifying any database. */
private fun readDatabases(context: Context): List<DatabaseData> = context.databaseList()
    .filter { !it.endsWith("-wal") && !it.endsWith("-shm") }
    .sorted()
    .mapNotNull { name ->
        val file = context.getDatabasePath(name)
        if (!file.exists()) return@mapNotNull null
        val db = SQLiteDatabase.openDatabase(file.path, null, SQLiteDatabase.OPEN_READONLY)
        try {
            val tables = mutableListOf<TableData>()
            db.rawQuery("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name", null).use { cursor ->
                while (cursor.moveToNext()) {
                    val tableName = cursor.getString(0)
                    val rows = mutableListOf<String>()
                    val quoted = "\"${tableName.replace("\"", "\"\"")}\""
                    db.rawQuery("SELECT * FROM $quoted", null).use { rowCursor ->
                        while (rowCursor.moveToNext()) {
                            val row = JSONObject()
                            for (i in 0 until rowCursor.columnCount) {
                                val value: Any? = when (rowCursor.getType(i)) {
                                    android.database.Cursor.FIELD_TYPE_NULL -> JSONObject.NULL
                                    android.database.Cursor.FIELD_TYPE_INTEGER -> rowCursor.getLong(i)
                                    android.database.Cursor.FIELD_TYPE_FLOAT -> rowCursor.getDouble(i)
                                    android.database.Cursor.FIELD_TYPE_BLOB -> "<BLOB ${rowCursor.getBlob(i).size} bytes>"
                                    else -> rowCursor.getString(i)
                                }
                                row.put(rowCursor.getColumnName(i), value)
                            }
                            rows += row.toString(2)
                        }
                    }
                    tables += TableData(tableName, rows)
                }
            }
            DatabaseData(name, tables)
        } finally {
            db.close()
        }
    }
