package com.chat.android.feature.reception

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import coil.compose.AsyncImage

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReceptionMenuScreen(
    id: String,
    code: String,
    codeType: String,
    onBack: () -> Unit,
    viewModel: ReceptionMenuViewModel = hiltViewModel()
) {
    val reception = viewModel.reception
    val orders = viewModel.receptionOrders
    val errorMessage = viewModel.errorMessage
    val isLoading = viewModel.isLoading

    LaunchedEffect(id, code, codeType) {
        viewModel.loadReception(id, code, codeType) { newType ->
            // In Vue it redirects to /3 if type was 2. 
            // Here we can handle navigation if needed, but for now we just reload.
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("メニュー一覧") },
                navigationIcon = {
                    TextButton(onClick = onBack) {
                        Text("戻る")
                    }
                }
            )
        }
    ) { padding ->
        if (isLoading && reception == null) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .padding(padding)
                    .fillMaxSize()
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(24.dp)
            ) {
                if (errorMessage != null) {
                    item {
                        Text(errorMessage, color = MaterialTheme.colorScheme.error)
                    }
                }

                if (reception != null) {
                    item {
                        Text(viewModel.seatName, style = MaterialTheme.typography.titleMedium)
                    }

                    items(reception.menus) { menu ->
                        MenuCard(
                            menu = menu,
                            viewModel = viewModel,
                            onAddToCart = { viewModel.addCart(menu) }
                        )
                    }

                    item {
                        Divider(Modifier.padding(vertical = 16.dp))
                        Text("カート内の注文 ${viewModel.seatName}", style = MaterialTheme.typography.headlineSmall)
                    }

                    if (orders.isEmpty()) {
                        item {
                            Text("カートに商品はありません。")
                        }
                    } else {
                        items(orders) { order ->
                            OrderCard(
                                order = order,
                                viewModel = viewModel,
                                onDelete = { viewModel.removeOrder(order.receptionOrderID) }
                            )
                        }

                        item {
                            Column(
                                modifier = Modifier.fillMaxWidth(),
                                horizontalAlignment = Alignment.End
                            ) {
                                Text(
                                    "合計: ${viewModel.totalPrice} 円",
                                    style = MaterialTheme.typography.titleLarge,
                                    fontWeight = FontWeight.Bold
                                )
                                Spacer(Modifier.height(16.dp))
                                Button(
                                    onClick = {
                                        viewModel.submitOrder(id, code) {
                                            // Order success
                                        }
                                    },
                                    modifier = Modifier.fillMaxWidth()
                                ) {
                                    Text("注文")
                                }
                            }
                        }
                    }
                } else if (!isLoading) {
                    item {
                        Text("メニューが見つかりません。")
                    }
                }
            }
        }
    }
}

@Composable
fun MenuCard(
    menu: Menu,
    viewModel: ReceptionMenuViewModel,
    onAddToCart: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        border = BorderStroke(1.dp, Color.LightGray)
    ) {
        Column(Modifier.padding(16.dp)) {
            Text(menu.menuName, style = MaterialTheme.typography.titleLarge)
            Text("価格: ${menu.price}円", style = MaterialTheme.typography.bodyMedium)
            Spacer(Modifier.height(8.dp))

            menu.items.forEach { itemID ->
                val item = viewModel.getItemDetail(itemID)
                if (item != null) {
                    Row(
                        modifier = Modifier.padding(vertical = 8.dp),
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        AsyncImage(
                            model = item.imgPath,
                            contentDescription = item.itemName,
                            modifier = Modifier.size(80.dp),
                            contentScale = ContentScale.Crop
                        )
                        Column {
                            Text(item.itemName, fontWeight = FontWeight.Bold)
                            item.choices.forEachIndexed { groupIndex, group ->
                                val key = "item-${item.itemID}-$groupIndex"
                                val selected = viewModel.selectedChoices[menu.menuID]?.get(key)
                                FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                                    group.forEach { choice ->
                                        Row(verticalAlignment = Alignment.CenterVertically) {
                                            RadioButton(
                                                selected = selected == choice,
                                                onClick = {
                                                    viewModel.selectedChoices[menu.menuID]?.set(key, choice)
                                                }
                                            )
                                            Text(choice, fontSize = 12.sp)
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }

            // Paid Options
            if (menu.paidOptions.isNotEmpty()) {
                Text("有料オプション:", style = MaterialTheme.typography.titleSmall)
                FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    menu.paidOptions.forEach { opt ->
                        val detail = viewModel.getItemDetail(opt.itemID)
                        OptionItem(
                            name = detail?.itemName ?: "",
                            price = opt.price,
                            image = detail?.imgPath,
                            selected = viewModel.selectedOptions[menu.menuID]?.paidOptions?.contains(opt.itemID) == true,
                            onToggle = {
                                val opts = viewModel.selectedOptions[menu.menuID]
                                if (opts != null) {
                                    if (opts.paidOptions.contains(opt.itemID)) {
                                        opts.paidOptions.remove(opt.itemID)
                                    } else {
                                        opts.paidOptions.add(opt.itemID)
                                    }
                                }
                            }
                        )
                    }
                }
            }

            // Free Options
            menu.freeOptions.forEachIndexed { patternIndex, optionArray ->
                Text("無料オプション:", style = MaterialTheme.typography.titleSmall)
                FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    optionArray.forEach { optionId ->
                        val detail = viewModel.getItemDetail(optionId)
                        OptionItem(
                            name = detail?.itemName ?: "",
                            image = detail?.imgPath,
                            selected = viewModel.selectedOptions[menu.menuID]?.freeOptions?.getOrNull(patternIndex) == optionId,
                            onToggle = {
                                val opts = viewModel.selectedOptions[menu.menuID]
                                if (opts != null && patternIndex < opts.freeOptions.size) {
                                    opts.freeOptions[patternIndex] = optionId
                                }
                            }
                        )
                    }
                }
            }

            // Free Multi Options
            if (menu.freeMultiOptions.isNotEmpty()) {
                Text("無料複数選択オプション:", style = MaterialTheme.typography.titleSmall)
                FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    menu.freeMultiOptions.forEach { optionId ->
                        val detail = viewModel.getItemDetail(optionId)
                        OptionItem(
                            name = detail?.itemName ?: "",
                            image = detail?.imgPath,
                            selected = viewModel.selectedOptions[menu.menuID]?.freeMultiOptions?.contains(optionId) == true,
                            onToggle = {
                                val opts = viewModel.selectedOptions[menu.menuID]
                                if (opts != null) {
                                    if (opts.freeMultiOptions.contains(optionId)) {
                                        opts.freeMultiOptions.remove(optionId)
                                    } else {
                                        opts.freeMultiOptions.add(optionId)
                                    }
                                }
                            }
                        )
                    }
                }
            }

            Spacer(Modifier.height(16.dp))
            Button(onClick = onAddToCart, modifier = Modifier.fillMaxWidth()) {
                Text("カートに追加")
            }
        }
    }
}

@Composable
fun OptionItem(
    name: String,
    price: Int? = null,
    image: String?,
    selected: Boolean,
    onToggle: () -> Unit
) {
    Card(
        modifier = Modifier
            .width(100.dp)
            .clickable { onToggle() },
        colors = CardDefaults.cardColors(
            containerColor = if (selected) Color(0xFFE6F0FF) else MaterialTheme.colorScheme.surface
        ),
        border = BorderStroke(2.dp, if (selected) Color(0xFF007BFF) else Color.Transparent)
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.padding(4.dp)
        ) {
            AsyncImage(
                model = image,
                contentDescription = name,
                modifier = Modifier
                    .size(80.dp)
                    .padding(4.dp),
                contentScale = ContentScale.Crop
            )
            Text(name, fontSize = 10.sp, maxLines = 1)
            if (price != null) {
                Text("${price}円", fontSize = 10.sp, color = Color.Gray)
            }
        }
    }
}

@Composable
fun OrderCard(
    order: ReceptionOrder,
    viewModel: ReceptionMenuViewModel,
    onDelete: () -> Unit
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        border = BorderStroke(1.dp, Color.LightGray)
    ) {
        Column(Modifier.padding(16.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(order.menuName, style = MaterialTheme.typography.titleMedium)
                IconButton(onClick = onDelete) {
                    Icon(Icons.Default.Close, contentDescription = "削除")
                }
            }

            order.items.forEach { item ->
                Row(verticalAlignment = Alignment.CenterVertically) {
                    AsyncImage(
                        model = item.imgPath,
                        contentDescription = item.itemName,
                        modifier = Modifier.size(60.dp),
                        contentScale = ContentScale.Crop
                    )
                    Spacer(Modifier.width(8.dp))
                    Column {
                        Text(item.itemName)
                        item.choices.forEachIndexed { groupIndex, _ ->
                            val key = "item-${item.itemID}-$groupIndex"
                            val choice = order.itemChoices[key]
                            if (choice != null) {
                                Text(choice, style = MaterialTheme.typography.bodySmall)
                            }
                        }
                    }
                }
            }

            if (order.freeOptions.isNotEmpty()) {
                Text("無料オプション:", style = MaterialTheme.typography.labelSmall)
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    order.freeOptions.forEach { id ->
                        val detail = viewModel.getItemDetail(id)
                        Text(detail?.itemName ?: id.toString(), style = MaterialTheme.typography.bodySmall)
                    }
                }
            }

            if (order.paidOptions.isNotEmpty()) {
                Text("有料オプション:", style = MaterialTheme.typography.labelSmall)
                order.paidOptions.forEach { opt ->
                    val detail = viewModel.getItemDetail(opt.itemID)
                    Text("${detail?.itemName ?: opt.itemID} ${opt.price}円", style = MaterialTheme.typography.bodySmall)
                }
            }

            if (order.freeMultiOptions.isNotEmpty()) {
                Text("複数選択オプション:", style = MaterialTheme.typography.labelSmall)
                order.freeMultiOptions.forEach { id ->
                    val detail = viewModel.getItemDetail(id)
                    Text(detail?.itemName ?: id.toString(), style = MaterialTheme.typography.bodySmall)
                }
            }

            Spacer(Modifier.height(8.dp))
            Text("メニュー価格: ${order.price}円", fontWeight = FontWeight.Bold)
        }
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun FlowRow(
    horizontalArrangement: Arrangement.Horizontal = Arrangement.Start,
    content: @Composable () -> Unit
) {
    androidx.compose.foundation.layout.FlowRow(
        horizontalArrangement = horizontalArrangement,
        content = { content() }
    )
}
