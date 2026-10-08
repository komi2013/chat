package com.chat.android.feature.reception

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.clickable
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.compose.ui.layout.ContentScale
import coil.compose.AsyncImage
import androidx.compose.foundation.Image
import androidx.compose.ui.res.painterResource
import com.chat.android.R
import androidx.compose.ui.window.Dialog
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.feature.channel.DbAlias
import com.chat.android.feature.channel.DbGroup
import androidx.compose.foundation.border
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.shape.RoundedCornerShape

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun ReceptionScreen(
    navController: NavController,
    id: String? = null,
    code: String? = null,
    viewModel: ReceptionViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsState()

    LaunchedEffect(id, code) {
        if (id != null) {
            viewModel.loadReception(id, code)
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("受付フォーム") },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                },
                actions = {
                    Button(
                        onClick = { viewModel.submit() },
                        enabled = !state.loading && !state.saving
                    ) {
                        Text("送信")
                    }
                }
            )
        }
    ) { padding ->
        if (state.loading) {
            Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else {
            val reception = state.reception
            LazyColumn(
                modifier = Modifier.fillMaxSize().padding(padding),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                if (state.error != null) {
                    item {
                        Text(state.error!!, color = MaterialTheme.colorScheme.error)
                    }
                }

                item {
                    OutlinedTextField(
                        value = reception.receptionTitle,
                        onValueChange = { viewModel.updateReception { copy(receptionTitle = it) } },
                        label = { Text("タイトル") },
                        modifier = Modifier.fillMaxWidth()
                    )
                }

                // 管理者
                item {
                    Text("管理者", style = MaterialTheme.typography.titleMedium)
                    SelectAlias(
                        selectedNames = reception.adminNames,
                        aliases = state.aliases,
                        groups = state.groups,
                        onAdd = { name ->
                            viewModel.updateReception { copy(adminNames = adminNames + name) }
                        },
                        onRemove = { name ->
                            viewModel.updateReception { copy(adminNames = adminNames.filter { it != name }) }
                        }
                    )
                }

                // 加入者
                item {
                    Text("加入者", style = MaterialTheme.typography.titleMedium)
                    SelectAlias(
                        selectedNames = reception.joinNames,
                        aliases = state.aliases,
                        groups = state.groups,
                        onAdd = { name ->
                            viewModel.updateReception { copy(joinNames = joinNames + name) }
                        },
                        onRemove = { name ->
                            viewModel.updateReception { copy(joinNames = joinNames.filter { it != name }) }
                        }
                    )
                }

                // スキル
                item {
                    Text("スキル", style = MaterialTheme.typography.titleMedium)
                }
                itemsIndexed(reception.skills) { index, skill ->
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        OutlinedTextField(
                            value = skill,
                            onValueChange = { viewModel.updateSkill(index, it) },
                            modifier = Modifier.weight(1f)
                        )
                        IconButton(onClick = { viewModel.removeSkill(index) }) {
                            Icon(Icons.Default.Delete, contentDescription = null)
                        }
                    }
                }
                item {
                    Button(onClick = { viewModel.addSkill() }) {
                        Icon(Icons.Default.Add, contentDescription = null)
                        Text("スキルを追加")
                    }
                }

                // メニュー
                item {
                    Text("メニュー", style = MaterialTheme.typography.titleMedium)
                }
                itemsIndexed(reception.menus) { index, menu ->
                    Card(modifier = Modifier.fillMaxWidth()) {
                        Column(modifier = Modifier.padding(8.dp)) {
                            Text("メニュー ID: ${menu.menuID}")
                            OutlinedTextField(
                                value = menu.menuName,
                                onValueChange = { newName ->
                                    viewModel.updateReception {
                                        copy(menus = menus.toMutableList().apply {
                                            set(index, menu.copy(menuName = newName))
                                        })
                                    }
                                },
                                label = { Text("メニュー名") },
                                modifier = Modifier.fillMaxWidth()
                            )
                            OutlinedTextField(
                                value = menu.price.toString(),
                                onValueChange = {
                                    val newPrice = it.toIntOrNull() ?: 0
                                    viewModel.updateReception {
                                        copy(menus = menus.toMutableList().apply {
                                            set(index, menu.copy(price = newPrice))
                                        })
                                    }
                                },
                                label = { Text("価格") },
                                modifier = Modifier.fillMaxWidth()
                            )
                            OutlinedTextField(
                                value = menu.prepaidPrice.toString(),
                                onValueChange = {
                                    val v = it.toIntOrNull() ?: 0
                                    viewModel.updateReception {
                                        copy(menus = menus.toMutableList().apply {
                                            set(index, menu.copy(prepaidPrice = v))
                                        })
                                    }
                                },
                                label = { Text("予約プリペイド価格") },
                                modifier = Modifier.fillMaxWidth()
                            )
                            // Need Skill Dropdown
                            var skillExpanded by remember { mutableStateOf(false) }
                            Box {
                                OutlinedTextField(
                                    value = menu.needSkill,
                                    onValueChange = {},
                                    label = { Text("必要スキル") },
                                    readOnly = true,
                                    modifier = Modifier.fillMaxWidth(),
                                    trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = skillExpanded) }
                                )
                                DropdownMenu(expanded = skillExpanded, onDismissRequest = { skillExpanded = false }) {
                                    reception.skills.forEach { skill ->
                                        DropdownMenuItem(
                                            text = { Text(skill) },
                                            onClick = {
                                                viewModel.updateReception {
                                                    copy(menus = menus.toMutableList().apply {
                                                        set(index, menu.copy(needSkill = skill))
                                                    })
                                                }
                                                skillExpanded = false
                                            }
                                        )
                                    }
                                }
                                Box(modifier = Modifier.matchParentSize().clickable { skillExpanded = true })
                            }
                            OutlinedTextField(
                                value = menu.needFacility,
                                onValueChange = { v ->
                                    viewModel.updateReception {
                                        copy(menus = menus.toMutableList().apply {
                                            set(index, menu.copy(needFacility = v))
                                        })
                                    }
                                },
                                label = { Text("必要設備") },
                                modifier = Modifier.fillMaxWidth()
                            )
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Checkbox(
                                    checked = menu.specifyNameFlag,
                                    onCheckedChange = { v ->
                                        viewModel.updateReception {
                                            copy(menus = menus.toMutableList().apply {
                                                set(index, menu.copy(specifyNameFlag = v))
                                            })
                                        }
                                    }
                                )
                                Text("指名フラグ")
                            }
                            OutlinedTextField(
                                value = menu.spendMinute.toString(),
                                onValueChange = {
                                    val v = it.toIntOrNull() ?: 0
                                    viewModel.updateReception {
                                        copy(menus = menus.toMutableList().apply {
                                            set(index, menu.copy(spendMinute = v))
                                        })
                                    }
                                },
                                label = { Text("所要時間(分)") },
                                modifier = Modifier.fillMaxWidth()
                            )

                            // Items selection
                            Text("商品の選択", style = MaterialTheme.typography.labelLarge)
                            menu.items.forEachIndexed { itemIdx, itemID ->
                                Row(verticalAlignment = Alignment.CenterVertically) {
                                    val itemName = reception.itemDetails.find { it.itemID == itemID }?.itemName ?: "未選択"
                                    Text(itemName, modifier = Modifier.weight(1f))
                                    IconButton(onClick = {
                                        viewModel.updateReception {
                                            copy(menus = menus.toMutableList().apply {
                                                set(index, menu.copy(items = menu.items.filterIndexed { i, _ -> i != itemIdx }))
                                            })
                                        }
                                    }) { Icon(Icons.Default.Delete, contentDescription = null) }
                                }
                            }
                            Button(onClick = {
                                // For simplicity, just add first item or 0
                                viewModel.updateReception {
                                    copy(menus = menus.toMutableList().apply {
                                        set(index, menu.copy(items = menu.items + 0))
                                    })
                                }
                            }) { Text("＋アイテム") }

                            // Paid Options
                            Text("有料オプション", style = MaterialTheme.typography.labelLarge)
                            menu.paidOptions.forEachIndexed { optIdx, opt ->
                                Row(verticalAlignment = Alignment.CenterVertically) {
                                    OutlinedTextField(
                                        value = opt.itemID.toString(),
                                        onValueChange = { v ->
                                            val id = v.toIntOrNull() ?: 0
                                            viewModel.updateReception {
                                                copy(menus = menus.toMutableList().apply {
                                                    val updatedOpts = menu.paidOptions.toMutableList().apply { set(optIdx, opt.copy(itemID = id)) }
                                                    set(index, menu.copy(paidOptions = updatedOpts))
                                                })
                                            }
                                        },
                                        label = { Text("アイテムID") },
                                        modifier = Modifier.width(80.dp)
                                    )
                                    Spacer(Modifier.width(8.dp))
                                    OutlinedTextField(
                                        value = opt.price.toString(),
                                        onValueChange = { v ->
                                            val p = v.toIntOrNull() ?: 0
                                            viewModel.updateReception {
                                                copy(menus = menus.toMutableList().apply {
                                                    val updatedOpts = menu.paidOptions.toMutableList().apply { set(optIdx, opt.copy(price = p)) }
                                                    set(index, menu.copy(paidOptions = updatedOpts))
                                                })
                                            }
                                        },
                                        label = { Text("価格") },
                                        modifier = Modifier.weight(1f)
                                    )
                                    IconButton(onClick = { viewModel.removeMenuPaidOption(index, optIdx) }) {
                                        Icon(Icons.Default.Delete, contentDescription = null)
                                    }
                                }
                            }
                            Button(onClick = { viewModel.addMenuPaidOption(index) }) { Text("＋有料オプション") }

                            // Free Options
                            Text("無料オプション(グループ)", style = MaterialTheme.typography.labelLarge)
                            menu.freeOptions.forEachIndexed { groupIdx, group ->
                                Column(modifier = Modifier.padding(start = 8.dp).border(1.dp, Color.Gray, RoundedCornerShape(4.dp)).padding(8.dp)) {
                                    Text("グループ ${groupIdx + 1}")
                                    group.forEachIndexed { itemIdx, itemID ->
                                        Row(verticalAlignment = Alignment.CenterVertically) {
                                            val itemName = reception.itemDetails.find { it.itemID == itemID }?.itemName ?: "ID: $itemID"
                                            Text(itemName, modifier = Modifier.weight(1f))
                                            IconButton(onClick = {
                                                viewModel.updateReception {
                                                    copy(menus = menus.toMutableList().apply {
                                                        val updatedGroups = menu.freeOptions.toMutableList().apply {
                                                            this[groupIdx] = this[groupIdx].filterIndexed { i, _ -> i != itemIdx }
                                                        }
                                                        set(index, menu.copy(freeOptions = updatedGroups))
                                                    })
                                                }
                                            }) { Icon(Icons.Default.Delete, contentDescription = null) }
                                        }
                                    }
                                    Button(onClick = {
                                        viewModel.updateReception {
                                            copy(menus = menus.toMutableList().apply {
                                                val updatedGroups = menu.freeOptions.toMutableList().apply {
                                                    this[groupIdx] = this[groupIdx] + 0
                                                }
                                                set(index, menu.copy(freeOptions = updatedGroups))
                                            })
                                        }
                                    }) { Text("＋アイテム") }
                                }
                            }
                            Button(onClick = { viewModel.addMenuFreeOptionGroup(index) }) { Text("＋無料オプション群") }

                            // Free Multi Options
                            Text("無料オプション(複数)", style = MaterialTheme.typography.labelLarge)
                            menu.freeMultiOptions.forEachIndexed { mIdx, itemID ->
                                Row(verticalAlignment = Alignment.CenterVertically) {
                                    val itemName = reception.itemDetails.find { it.itemID == itemID }?.itemName ?: "ID: $itemID"
                                    Text(itemName, modifier = Modifier.weight(1f))
                                    IconButton(onClick = {
                                        viewModel.updateReception {
                                            copy(menus = menus.toMutableList().apply {
                                                set(index, menu.copy(freeMultiOptions = menu.freeMultiOptions.filterIndexed { i, _ -> i != mIdx }))
                                            })
                                        }
                                    }) { Icon(Icons.Default.Delete, contentDescription = null) }
                                }
                            }
                            Button(onClick = {
                                viewModel.updateReception {
                                    copy(menus = menus.toMutableList().apply {
                                        set(index, menu.copy(freeMultiOptions = menu.freeMultiOptions + 0))
                                    })
                                }
                            }) { Text("＋アイテム") }

                            IconButton(onClick = { viewModel.removeMenu(index) }) {
                                Icon(Icons.Default.Delete, contentDescription = null)
                            }
                        }
                    }
                }
                item {
                    Button(onClick = { viewModel.addMenu() }) {
                        Icon(Icons.Default.Add, contentDescription = null)
                        Text("メニューを追加")
                    }
                }

                // アイテム詳細
                item {
                    Text("アイテム詳細", style = MaterialTheme.typography.titleMedium)
                }
                itemsIndexed(reception.itemDetails) { index, item ->
                    Card(modifier = Modifier.fillMaxWidth()) {
                        Column(modifier = Modifier.padding(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                OutlinedTextField(
                                    value = item.itemID.toString(),
                                    onValueChange = { val id = it.toIntOrNull() ?: 0
                                        viewModel.updateReception { copy(itemDetails = itemDetails.toMutableList().apply { set(index, item.copy(itemID = id)) }) }
                                    },
                                    label = { Text("ID") },
                                    modifier = Modifier.width(80.dp)
                                )
                                Spacer(Modifier.width(8.dp))
                                OutlinedTextField(
                                    value = item.itemName,
                                    onValueChange = { name ->
                                        viewModel.updateReception { copy(itemDetails = itemDetails.toMutableList().apply { set(index, item.copy(itemName = name)) }) }
                                    },
                                    label = { Text("名前") },
                                    modifier = Modifier.weight(1f)
                                )
                                IconButton(onClick = { viewModel.removeItemDetail(index) }) {
                                    Icon(Icons.Default.Delete, contentDescription = null)
                                }
                            }
                            // Item Image Path
                            Row(verticalAlignment = Alignment.CenterVertically, modifier = Modifier.fillMaxWidth()) {
                                if (item.imgPath.isNotEmpty()) {
                                    AsyncImage(
                                        model = item.imgPath,
                                        contentDescription = null,
                                        modifier = Modifier.size(64.dp).padding(end = 8.dp),
                                        contentScale = ContentScale.Crop
                                    )
                                }
                                OutlinedTextField(
                                    value = item.imgPath,
                                    onValueChange = { path ->
                                        viewModel.updateReception { copy(itemDetails = itemDetails.toMutableList().apply { set(index, item.copy(imgPath = path)) }) }
                                    },
                                    label = { Text("画像パス / URL") },
                                    modifier = Modifier.weight(1f)
                                )
                            }

                            // choices (二重配列)
                            Text("利用方法", style = MaterialTheme.typography.labelLarge)
                            item.choices.forEachIndexed { groupIndex, group ->
                                Column(modifier = Modifier.padding(start = 8.dp).border(1.dp, Color.Gray, RoundedCornerShape(4.dp)).padding(8.dp)) {
                                    Text("グループ ${groupIndex + 1}")
                                    group.forEachIndexed { choiceIndex, choice ->
                                        Row(verticalAlignment = Alignment.CenterVertically) {
                                            OutlinedTextField(
                                                value = choice,
                                                onValueChange = { newChoice ->
                                                    viewModel.updateReception {
                                                        copy(itemDetails = itemDetails.toMutableList().apply {
                                                            val updatedChoices = item.choices.toMutableList().apply {
                                                                this[groupIndex] = this[groupIndex].toMutableList().apply { set(choiceIndex, newChoice) }
                                                            }
                                                            set(index, item.copy(choices = updatedChoices))
                                                        })
                                                    }
                                                },
                                                modifier = Modifier.weight(1f)
                                            )
                                            IconButton(onClick = {
                                                viewModel.updateReception {
                                                    copy(itemDetails = itemDetails.toMutableList().apply {
                                                        val updatedChoices = item.choices.toMutableList().apply {
                                                            this[groupIndex] = this[groupIndex].filterIndexed { ci, _ -> ci != choiceIndex }
                                                        }
                                                        set(index, item.copy(choices = updatedChoices))
                                                    })
                                                }
                                            }) { Icon(Icons.Default.Delete, contentDescription = null) }
                                        }
                                    }
                                    Button(onClick = { viewModel.addItemDetailChoice(index, groupIndex) }) {
                                        Text("＋選択肢")
                                    }
                                }
                            }
                            Button(onClick = { viewModel.addItemDetailChoiceGroup(index) }) {
                                Text("＋利用方法追加")
                            }
                        }
                    }
                }
                item {
                    Button(onClick = { viewModel.addItemDetail() }) {
                        Icon(Icons.Default.Add, contentDescription = null)
                        Text("アイテムを追加")
                    }
                }

                // 待機設定
                item {
                    Text("待機設定", style = MaterialTheme.typography.titleMedium)
                }
                itemsIndexed(reception.waitConfigs) { index, config ->
                    Card(modifier = Modifier.fillMaxWidth()) {
                        Column(modifier = Modifier.padding(8.dp)) {
                            Row {
                                OutlinedTextField(
                                    value = config.guestRange.getOrElse(0) { 1 }.toString(),
                                    onValueChange = { val v = it.toIntOrNull() ?: 1
                                        viewModel.updateReception { copy(waitConfigs = waitConfigs.toMutableList().apply { set(index, config.copy(guestRange = listOf(v, config.guestRange.getOrElse(1) { 5 }))) }) }
                                    },
                                    label = { Text("最小人数") },
                                    modifier = Modifier.weight(1f)
                                )
                                Spacer(Modifier.width(8.dp))
                                OutlinedTextField(
                                    value = config.guestRange.getOrElse(1) { 5 }.toString(),
                                    onValueChange = { val v = it.toIntOrNull() ?: 5
                                        viewModel.updateReception { copy(waitConfigs = waitConfigs.toMutableList().apply { set(index, config.copy(guestRange = listOf(config.guestRange.getOrElse(0) { 1 }, v))) }) }
                                    },
                                    label = { Text("最大人数") },
                                    modifier = Modifier.weight(1f)
                                )
                            }
                            OutlinedTextField(
                                value = config.waitRatio.toString(),
                                onValueChange = { val v = it.toIntOrNull() ?: 0
                                    viewModel.updateReception { copy(waitConfigs = waitConfigs.toMutableList().apply { set(index, config.copy(waitRatio = v)) }) }
                                },
                                label = { Text("待機単位(分)") },
                                modifier = Modifier.fillMaxWidth()
                            )
                            IconButton(onClick = {
                                viewModel.updateReception { copy(waitConfigs = waitConfigs.filterIndexed { i, _ -> i != index }) }
                            }) { Icon(Icons.Default.Delete, contentDescription = null) }
                        }
                    }
                }
                item {
                    Button(onClick = { viewModel.addWaitConfig() }) {
                        Icon(Icons.Default.Add, contentDescription = null)
                        Text("待機設定を追加")
                    }
                }

                // パスコード
                item {
                    Text("パスコード", style = MaterialTheme.typography.titleMedium)
                }
                itemsIndexed(reception.passcodes) { index, passcode ->
                    Card(modifier = Modifier.fillMaxWidth()) {
                        Column(modifier = Modifier.padding(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            OutlinedTextField(
                                value = passcode.passkey,
                                onValueChange = { v -> viewModel.updateReception { copy(passcodes = passcodes.toMutableList().apply { set(index, passcode.copy(passkey = v)) }) } },
                                label = { Text("パスキー") },
                                modifier = Modifier.fillMaxWidth()
                            )
                            OutlinedTextField(
                                value = passcode.usageLimit.toString(),
                                onValueChange = { v -> val i = v.toIntOrNull() ?: 1
                                    viewModel.updateReception { copy(passcodes = passcodes.toMutableList().apply { set(index, passcode.copy(usageLimit = i)) }) }
                                },
                                label = { Text("使用上限") },
                                modifier = Modifier.fillMaxWidth()
                            )
                            Row {
                                OutlinedTextField(
                                    value = passcode.passStart,
                                    onValueChange = { v -> viewModel.updateReception { copy(passcodes = passcodes.toMutableList().apply { set(index, passcode.copy(passStart = v)) }) } },
                                    label = { Text("開始") },
                                    modifier = Modifier.weight(1f)
                                )
                                Spacer(Modifier.width(8.dp))
                                OutlinedTextField(
                                    value = passcode.passEnd,
                                    onValueChange = { v -> viewModel.updateReception { copy(passcodes = passcodes.toMutableList().apply { set(index, passcode.copy(passEnd = v)) }) } },
                                    label = { Text("終了") },
                                    modifier = Modifier.weight(1f)
                                )
                            }
                            IconButton(onClick = { viewModel.removePasscode(index) }) {
                                Icon(Icons.Default.Delete, contentDescription = null)
                            }
                        }
                    }
                }
                item {
                    Button(onClick = { viewModel.addPasscode() }) {
                        Icon(Icons.Default.Add, contentDescription = null)
                        Text("パスコードを追加")
                    }
                }

                // 質問 (Asks)
                item {
                    AskSection("質問 (単数回答)", reception.asks, { viewModel.addAsk() }, { idx -> viewModel.removeAsk(idx) }) { idx, ask ->
                        AskItem(ask, onUpdate = { updated ->
                            viewModel.updateReception { copy(asks = asks.toMutableList().apply { set(idx, updated) }) }
                        })
                    }
                }

                item {
                    AskSection("質問 (選択肢)", reception.askChoices, { viewModel.addAskChoice() }, { idx -> viewModel.removeAskChoice(idx) }) { idx, ask ->
                        AskItem(ask, showChoices = true, onUpdate = { updated ->
                            viewModel.updateReception { copy(askChoices = askChoices.toMutableList().apply { set(idx, updated) }) }
                        })
                    }
                }

                item {
                    AskSection("質問 (複数回答)", reception.askMultiChoices, { viewModel.addAskMultiChoice() }, { idx -> viewModel.removeAskMultiChoice(idx) }) { idx, ask ->
                        AskItem(ask, showChoices = true, onUpdate = { updated ->
                            viewModel.updateReception { copy(askMultiChoices = askMultiChoices.toMutableList().apply { set(idx, updated) }) }
                        })
                    }
                }

                // 営業時間
                item {
                    Text("営業時間", style = MaterialTheme.typography.titleMedium)
                }
                itemsIndexed(reception.openTimes) { index, time ->
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        OutlinedTextField(
                            value = time.limitStart,
                            onValueChange = { v -> viewModel.updateReception { copy(openTimes = openTimes.toMutableList().apply { set(index, time.copy(limitStart = v)) }) } },
                            label = { Text("開始") },
                            modifier = Modifier.weight(1f)
                        )
                        Spacer(Modifier.width(8.dp))
                        OutlinedTextField(
                            value = time.limitEnd,
                            onValueChange = { v -> viewModel.updateReception { copy(openTimes = openTimes.toMutableList().apply { set(index, time.copy(limitEnd = v)) }) } },
                            label = { Text("終了") },
                            modifier = Modifier.weight(1f)
                        )
                        IconButton(onClick = { viewModel.removeOpenTime(index) }) {
                            Icon(Icons.Default.Delete, contentDescription = null)
                        }
                    }
                }
                item {
                    Button(onClick = { viewModel.addOpenTime() }) {
                        Icon(Icons.Default.Add, contentDescription = null)
                        Text("営業時間追加")
                    }
                }

                // シフト
                item {
                    Text("シフト", style = MaterialTheme.typography.titleMedium)
                }
                itemsIndexed(reception.shifts) { index, shift ->
                    Card(modifier = Modifier.fillMaxWidth()) {
                        Column(modifier = Modifier.padding(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            SelectAlias(
                                selectedNames = shift.aliasNames,
                                aliases = state.aliases,
                                groups = state.groups,
                                onAdd = { name ->
                                    viewModel.updateReception { copy(shifts = shifts.toMutableList().apply { set(index, shift.copy(aliasNames = shift.aliasNames + name)) }) }
                                },
                                onRemove = { name ->
                                    viewModel.updateReception { copy(shifts = shifts.toMutableList().apply { set(index, shift.copy(aliasNames = shift.aliasNames.filter { it != name })) }) }
                                }
                            )
                            Row {
                                OutlinedTextField(
                                    value = shift.shiftStart,
                                    onValueChange = { v -> viewModel.updateReception { copy(shifts = shifts.toMutableList().apply { set(index, shift.copy(shiftStart = v)) }) } },
                                    label = { Text("開始") },
                                    modifier = Modifier.weight(1f)
                                )
                                Spacer(Modifier.width(8.dp))
                                OutlinedTextField(
                                    value = shift.shiftEnd,
                                    onValueChange = { v -> viewModel.updateReception { copy(shifts = shifts.toMutableList().apply { set(index, shift.copy(shiftEnd = v)) }) } },
                                    label = { Text("終了") },
                                    modifier = Modifier.weight(1f)
                                )
                            }
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Text("公開数:")
                                Slider(
                                    value = shift.open.toFloat(),
                                    onValueChange = { v -> viewModel.updateReception { copy(shifts = shifts.toMutableList().apply { set(index, shift.copy(open = v.toInt())) }) } },
                                    valueRange = 0f..50f,
                                    modifier = Modifier.weight(1f)
                                )
                                Text(shift.open.toString())
                            }
                            IconButton(onClick = { viewModel.removeShift(index) }) {
                                Icon(Icons.Default.Delete, contentDescription = null)
                            }
                        }
                    }
                }
                item {
                    Button(onClick = { viewModel.addShift() }) {
                        Icon(Icons.Default.Add, contentDescription = null)
                        Text("シフト追加")
                    }
                }

                // 設備
                item {
                    Text("設備", style = MaterialTheme.typography.titleMedium)
                }
                itemsIndexed(reception.facilities) { index, facility ->
                    Card(modifier = Modifier.fillMaxWidth()) {
                        Column(modifier = Modifier.padding(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            OutlinedTextField(
                                value = facility.facilityName,
                                onValueChange = { v -> viewModel.updateReception { copy(facilities = facilities.toMutableList().apply { set(index, facility.copy(facilityName = v)) }) } },
                                label = { Text("設備名") },
                                modifier = Modifier.fillMaxWidth()
                            )
                            OutlinedTextField(
                                value = facility.capacity.toString(),
                                onValueChange = { v -> val i = v.toIntOrNull() ?: 0
                                    viewModel.updateReception { copy(facilities = facilities.toMutableList().apply { set(index, facility.copy(capacity = i)) }) }
                                },
                                label = { Text("収容人数") },
                                modifier = Modifier.fillMaxWidth()
                            )
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                Checkbox(
                                    checked = facility.bookable,
                                    onCheckedChange = { v -> viewModel.updateReception { copy(facilities = facilities.toMutableList().apply { set(index, facility.copy(bookable = v)) }) } }
                                )
                                Text("予約可能")
                            }
                            IconButton(onClick = { viewModel.removeFacility(index) }) {
                                Icon(Icons.Default.Delete, contentDescription = null)
                            }
                        }
                    }
                }
                item {
                    Button(onClick = { viewModel.addFacility() }) {
                        Icon(Icons.Default.Add, contentDescription = null)
                        Text("設備追加")
                    }
                }

                // スタッフスキル
                item {
                    Text("スタッフスキル", style = MaterialTheme.typography.titleMedium)
                }
                itemsIndexed(reception.staffSkills) { index, staffSkill ->
                    Card(modifier = Modifier.fillMaxWidth()) {
                        Column(modifier = Modifier.padding(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            SelectAlias(
                                selectedNames = if (staffSkill.aliasName.isEmpty()) emptyList() else listOf(staffSkill.aliasName),
                                aliases = state.aliases,
                                groups = state.groups,
                                onAdd = { name -> viewModel.updateReception { copy(staffSkills = staffSkills.toMutableList().apply { set(index, staffSkill.copy(aliasName = name)) }) } },
                                onRemove = { _ -> viewModel.updateReception { copy(staffSkills = staffSkills.toMutableList().apply { set(index, staffSkill.copy(aliasName = "")) }) } }
                            )
                            
                            FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                staffSkill.skills.forEachIndexed { sIdx, skill ->
                                    var expanded by remember { mutableStateOf(false) }
                                    Box {
                                        InputChip(
                                            selected = false,
                                            onClick = { expanded = true },
                                            label = { Text(skill.ifEmpty { "スキル選択" }) },
                                            trailingIcon = {
                                                Icon(Icons.Default.Delete, contentDescription = null, modifier = Modifier.size(16.dp).clickable {
                                                    viewModel.removeStaffSkillEntry(index, sIdx)
                                                })
                                            }
                                        )
                                        DropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
                                            reception.skills.forEach { s ->
                                                DropdownMenuItem(
                                                    text = { Text(s) },
                                                    onClick = {
                                                        viewModel.updateReception {
                                                            copy(staffSkills = staffSkills.toMutableList().apply {
                                                                val updatedSkills = staffSkill.skills.toMutableList().apply { set(sIdx, s) }
                                                                set(index, staffSkill.copy(skills = updatedSkills))
                                                            })
                                                        }
                                                        expanded = false
                                                    }
                                                )
                                            }
                                        }
                                    }
                                }
                                AssistChip(
                                    onClick = { viewModel.addStaffSkillEntry(index) },
                                    label = { Text("＋スキル") }
                                )
                            }

                            IconButton(onClick = { viewModel.removeStaffSkill(index) }) {
                                Icon(Icons.Default.Delete, contentDescription = null)
                            }
                        }
                    }
                }
                item {
                    Button(onClick = { viewModel.addStaffSkill() }) {
                        Icon(Icons.Default.Add, contentDescription = null)
                        Text("スタッフスキル追加")
                    }
                }

                item {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Checkbox(
                            checked = reception.workStaffNeed,
                            onCheckedChange = { v -> viewModel.updateReception { copy(workStaffNeed = v) } }
                        )
                        Text("スタッフ募集を行う")
                    }
                }
            }
        }
    }
}

@Composable
fun <T> AskSection(
    title: String,
    items: List<T>,
    onAdd: () -> Unit,
    onRemove: (Int) -> Unit,
    content: @Composable (Int, T) -> Unit
) {
    Column {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Text(title, style = MaterialTheme.typography.titleMedium, modifier = Modifier.weight(1f))
            Button(onClick = onAdd) {
                Icon(Icons.Default.Add, contentDescription = null)
                Text("追加")
            }
        }
        items.forEachIndexed { index, item ->
            Row(verticalAlignment = Alignment.CenterVertically) {
                Box(modifier = Modifier.weight(1f)) {
                    content(index, item)
                }
                IconButton(onClick = { onRemove(index) }) {
                    Icon(Icons.Default.Delete, contentDescription = null)
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun AskItem(
    ask: Ask,
    showChoices: Boolean = false,
    onUpdate: (Ask) -> Unit
) {
    Column(modifier = Modifier.padding(vertical = 4.dp)) {
        OutlinedTextField(
            value = ask.question,
            onValueChange = { onUpdate(ask.copy(question = it)) },
            label = { Text("質問文") },
            modifier = Modifier.fillMaxWidth()
        )
        if (showChoices) {
            FlowRow(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                ask.choices.forEachIndexed { ci, choice ->
                    InputChip(
                        selected = false,
                        onClick = {},
                        label = {
                            BasicTextField(
                                value = choice,
                                onValueChange = { v: String ->
                                    val newChoices = ask.choices.toMutableList().apply { set(ci, v) }
                                    onUpdate(ask.copy(choices = newChoices))
                                }
                            )
                        },
                        trailingIcon = {
                            Icon(
                                Icons.Default.Delete,
                                contentDescription = null,
                                modifier = Modifier.size(16.dp).clickable {
                                    onUpdate(ask.copy(choices = ask.choices.filterIndexed { i, _ -> i != ci }))
                                }
                            )
                        }
                    )
                }
                AssistChip(
                    onClick = { onUpdate(ask.copy(choices = ask.choices + "")) },
                    label = { Text("＋選択肢") }
                )
            }
        }
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun SelectAlias(
    selectedNames: List<String>,
    aliases: List<DbAlias>,
    groups: List<DbGroup>,
    onAdd: (String) -> Unit,
    onRemove: (String) -> Unit
) {
    var showDialog by remember { mutableStateOf(false) }
    var search by remember { mutableStateOf("") }

    Column {
        FlowRow(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            selectedNames.forEach { name ->
                AssistChip(
                    onClick = { onRemove(name) },
                    label = { Text(name) },
                    trailingIcon = { Icon(Icons.Default.Delete, contentDescription = null, modifier = Modifier.size(16.dp)) }
                )
            }
            AssistChip(
                onClick = { showDialog = true },
                label = { Text("追加") },
                leadingIcon = { Icon(Icons.Default.Add, contentDescription = null, modifier = Modifier.size(16.dp)) }
            )
        }
    }

    if (showDialog) {
        AlertDialog(
            onDismissRequest = { showDialog = false },
            title = { Text("メンバーを選択") },
            text = {
                Column {
                    OutlinedTextField(
                        value = search,
                        onValueChange = { search = it },
                        label = { Text("検索") },
                        modifier = Modifier.fillMaxWidth()
                    )
                    Spacer(Modifier.height(8.dp))
                    LazyColumn(modifier = Modifier.heightIn(max = 300.dp)) {
                        val filteredAliases = aliases.filter {
                            it.aliasName.contains(search, ignoreCase = true) && it.aliasName !in selectedNames
                        }
                        items(filteredAliases) { alias ->
                            ListItem(
                                headlineContent = { Text(alias.aliasName) },
                                leadingContent = {
                                    if (alias.aliasImg.isNotEmpty()) {
                                        AsyncImage(
                                            model = alias.aliasImg,
                                            contentDescription = null,
                                            modifier = Modifier.size(40.dp),
                                            contentScale = ContentScale.Crop
                                        )
                                    } else {
                                        Icon(Icons.Default.Refresh, contentDescription = null)
                                    }
                                },
                                modifier = Modifier.clickable {
                                    onAdd(alias.aliasName)
                                    showDialog = false
                                }
                            )
                        }
                    }
                }
            },
            confirmButton = {
                TextButton(onClick = { showDialog = false }) { Text("閉じる") }
            }
        )
    }
}
