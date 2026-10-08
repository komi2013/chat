package com.chat.android.feature.reception

import androidx.compose.foundation.Image
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import coil.compose.rememberAsyncImagePainter

@Composable
fun ReceptionOrderScreen(
    id: String,
    code: String,
    viewModel: ReceptionOrderViewModel = hiltViewModel()
) {
    LaunchedEffect(id, code) {
        viewModel.init(id, code)
    }

    val orders = viewModel.receptionOrders
    val totalPrice = viewModel.totalPrice
    val seatName = viewModel.seatName
    val iamStaff = viewModel.iamStaff
    val errorMessage = viewModel.errorMessage

    Scaffold(
        topBar = {
            SmallTopAppBar(title = { Text("注文履歴") })
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .padding(padding)
                .fillMaxSize()
                .padding(16.dp)
        ) {
            if (errorMessage != null) {
                Text(
                    text = errorMessage,
                    color = Color.Red,
                    modifier = Modifier.padding(bottom = 16.dp)
                )
            }

            Text(
                text = "カート内の注文 $seatName",
                style = MaterialTheme.typography.titleLarge,
                modifier = Modifier.padding(bottom = 16.dp)
            )

            if (orders.isEmpty()) {
                Box(modifier = Modifier.weight(1f), contentAlignment = Alignment.Center) {
                    Text("カートに商品はありません。")
                }
            } else {
                LazyColumn(modifier = Modifier.weight(1f)) {
                    items(orders) { order ->
                        OrderItemView(order, iamStaff, viewModel::getItemDetail)
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            Text(
                text = "合計: $totalPrice 円",
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold
            )

            if (iamStaff) {
                Button(
                    onClick = { viewModel.deleteOrder(id, code) },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 16.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = Color.Red)
                ) {
                    Text("会計終了", color = Color.White)
                }
            }
        }
    }
}

@Composable
fun OrderItemView(
    order: ReceptionOrder,
    iamStaff: Boolean,
    getItemDetail: (Int) -> ItemDetail?
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 8.dp)
            .border(1.dp, Color.LightGray, RoundedCornerShape(8.dp)),
        colors = CardDefaults.cardColors(containerColor = Color.White)
    ) {
        Column(modifier = Modifier.padding(12.dp)) {
            Text(text = order.menuName, style = MaterialTheme.typography.titleMedium)
            
            if (iamStaff) {
                Text(
                    text = "注文ID: ${order.receptionOrderID}",
                    style = MaterialTheme.typography.bodySmall,
                    color = Color.Gray
                )
            }

            Spacer(modifier = Modifier.height(8.dp))

            // Items
            order.items.forEach { item ->
                Row(verticalAlignment = Alignment.CenterVertically) {
                    if (!item.imgPath.isNullOrEmpty()) {
                        Image(
                            painter = rememberAsyncImagePainter(item.imgPath),
                            contentDescription = item.itemName,
                            modifier = Modifier.size(40.dp),
                            contentScale = ContentScale.Crop
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                    }
                    Text(text = item.itemName)
                }

                // Choices
                if (item.choices.isNotEmpty()) {
                    item.choices.forEachIndexed { groupIndex, _ ->
                        val choice = order.itemChoices["item-${item.itemID}-$groupIndex"]
                        if (choice != null) {
                            Text(
                                text = choice,
                                style = MaterialTheme.typography.bodySmall,
                                modifier = Modifier.padding(start = 48.dp)
                            )
                        }
                    }
                }
            }

            // Options
            if (order.freeOptions.isNotEmpty()) {
                OptionSection(title = "無料オプション", ids = order.freeOptions, getItemDetail = getItemDetail)
            }

            if (order.paidOptions.isNotEmpty()) {
                Text(
                    text = "有料オプション:",
                    style = MaterialTheme.typography.bodySmall,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.padding(top = 8.dp)
                )
                order.paidOptions.forEach { opt ->
                    val detail = getItemDetail(opt.itemID)
                    Row(verticalAlignment = Alignment.CenterVertically, modifier = Modifier.padding(start = 8.dp)) {
                        if (!detail?.imgPath.isNullOrEmpty()) {
                            Image(
                                painter = rememberAsyncImagePainter(detail?.imgPath),
                                contentDescription = detail?.itemName,
                                modifier = Modifier.size(30.dp),
                                contentScale = ContentScale.Crop
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                        }
                        Text(
                            text = "${detail?.itemName ?: ""} (+${opt.price}円)",
                            style = MaterialTheme.typography.bodySmall
                        )
                    }
                }
            }

            if (order.freeMultiOptions.isNotEmpty()) {
                OptionSection(title = "複数選択オプション", ids = order.freeMultiOptions, getItemDetail = getItemDetail)
            }

            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "メニュー価格: ${order.price}円",
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.align(Alignment.End)
            )
        }
    }
}

@Composable
fun OptionSection(title: String, ids: List<Int>, getItemDetail: (Int) -> ItemDetail?) {
    Text(
        text = "$title:",
        style = MaterialTheme.typography.bodySmall,
        fontWeight = FontWeight.Bold,
        modifier = Modifier.padding(top = 8.dp)
    )
    ids.forEach { id ->
        val detail = getItemDetail(id)
        Row(verticalAlignment = Alignment.CenterVertically, modifier = Modifier.padding(start = 8.dp)) {
            if (!detail?.imgPath.isNullOrEmpty()) {
                Image(
                    painter = rememberAsyncImagePainter(detail?.imgPath),
                    contentDescription = detail?.itemName,
                    modifier = Modifier.size(30.dp),
                    contentScale = ContentScale.Crop
                )
                Spacer(modifier = Modifier.width(8.dp))
            }
            Text(text = detail?.itemName ?: "", style = MaterialTheme.typography.bodySmall)
        }
    }
}
