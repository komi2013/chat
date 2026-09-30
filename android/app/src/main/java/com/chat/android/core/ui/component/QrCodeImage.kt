package com.chat.android.core.ui.component

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.size
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import com.chat.android.core.util.QrEncoder

/**
 * QRコードを表示する。
 *
 * Maven Central に到達できない環境でも招待QRを出せるよう、
 * [QrEncoder] で自前生成したモジュール配列を Canvas に描く。
 * 仕様どおり4モジュール分の余白（quiet zone）を確保する。
 */
@Composable
fun QrCodeImage(
    content: String,
    modifier: Modifier = Modifier,
    size: Dp = 260.dp,
    darkColor: Color = Color.Black,
    lightColor: Color = Color.White,
    quietZoneModules: Int = 4
) {
    // 文字列が変わらない限り再生成しない
    val qr = remember(content) { QrEncoder.encode(content) }

    if (qr == null) {
        // 容量超過などで生成できなかった場合は何も描かない
        Box(modifier.size(size))
        return
    }

    val totalModules = qr.size + quietZoneModules * 2

    // Canvas の描画座標は dp ではなく px である。
    // moduleSize.value（dp）を使って描画すると、密度 2.0 の端末では
    // 240dp=480px のキャンバスに対して 1モジュールが約 1/2 倍の寸法になり、
    // QRが左上に小さく固まって見えてしまう。
    // そのためキャンバスの実サイズ(px)からモジュール寸分を求める。
    Canvas(modifier = modifier.size(size)) {
        val canvasSize = this.size.minDimension
        val modulePx = canvasSize / totalModules

        // 余白（lightColor で塗る）
        drawRect(color = lightColor, size = Size(this.size.width, this.size.height))

        for (y in 0 until qr.size) {
            for (x in 0 until qr.size) {
                if (!qr.modules[y][x]) continue
                val left = (x + quietZoneModules) * modulePx
                val top = (y + quietZoneModules) * modulePx
                drawRect(
                    color = darkColor,
                    topLeft = Offset(left, top),
                    size = Size(modulePx, modulePx)
                )
            }
        }
    }
}

@Preview(showBackground = true)
@Composable
private fun QrCodeImagePreview() {
    QrCodeImage(
        content = "https://chat.quigen.info/profile/J/?code=DtG9LRpT39W7Xcd9",
        size = 220.dp,
        modifier = Modifier.background(Color.White)
    )
}
