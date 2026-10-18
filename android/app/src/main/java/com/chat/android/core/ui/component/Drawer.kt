package com.chat.android.core.ui.component

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import androidx.navigation.NavController
import com.chat.android.core.repository.UserRepository
import com.chat.android.feature.channel.ChannelRepository
import com.chat.android.navigation.LoginRoute
import com.chat.android.navigation.UserRoute
import com.chat.android.navigation.SettingsRoute
import com.chat.android.navigation.EntryFormListRoute
import com.chat.android.navigation.ChannelRoute
import com.chat.android.navigation.HomeRoute
import com.chat.android.navigation.CalendarRoute
import com.chat.android.navigation.TimestampCodeRoute
import com.chat.android.navigation.ReceptionRoute
import com.chat.android.navigation.TicketListRoute
import javax.inject.Inject
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

data class DrawerUiState(
    val hasChannel: Boolean = false,
    val channelName: String = "",
    val myAlias: String = "",
    val isSignedIn: Boolean = false
)

/**
 * ドロワーの表示状態。
 *
 * 注意点: 以前は AppDrawer のデフォルト引数で
 * `DrawerViewModel(hiltViewModel<UserViewModel>().userRepository)` と生成していたため、
 * 再コンポジションのたびに新しい mutableStateOf を持つインスタンスが作られていた。
 * その結果 Compose が保持する古いスナップショット（サインイン前の値）が読み出され、
 * サインイン済みでも「ユーザー設定」「チャネル登録」「フォーム編集」が
 * 一覧に出ない不具合があった。Hilt の @HiltViewModel にして hiltViewModel() で取得する。
 */
@HiltViewModel
class DrawerViewModel @Inject constructor(
    private val userRepository: UserRepository,
    private val channelRepository: ChannelRepository
) : ViewModel() {
    private val _uiState = MutableStateFlow(DrawerUiState())
    val uiState = _uiState.asStateFlow()

    init {
        // サインイン状態の変更を監視
        viewModelScope.launch {
            userRepository.signInState.collect { signedIn ->
                _uiState.update { it.copy(isSignedIn = signedIn) }
            }
        }
        // チャネル選択の変化だけ追って、ヘッダー表示を更新する。
        viewModelScope.launch {
            userRepository.currentChannelId.collect { channelId ->
                val loaded = loadInto(_uiState.value, channelId)
                _uiState.update { it.copy(
                    hasChannel = loaded.hasChannel,
                    channelName = loaded.channelName,
                    myAlias = loaded.myAlias
                ) }
            }
        }
        // エイリアス切替・チャネル編集はローカルDBを書き換えるので、ヘッダー表示も更新する。
        viewModelScope.launch {
            channelRepository.dbUpdateFlow.collect {
                val loaded = loadInto(_uiState.value, userRepository.currentChannelId.value)
                _uiState.update { it.copy(
                    hasChannel = loaded.hasChannel,
                    channelName = loaded.channelName,
                    myAlias = loaded.myAlias
                ) }
            }
        }
    }

    /**
     * チャネル/スレッドを開いたときに「現在のチャネル」を記録する。
     * vue の Channel.vue / Thread.vue の `localStorage.setItem('channelID', ...)` と同じ。
     */
    fun onChannelOpened(channelId: String) {
        if (channelId.isNotBlank()) userRepository.setCurrentChannelId(channelId)
    }

    /**
     * 「現在のチャネル」からチャネル名・自分のエイリアスを読み込む。
     * channelID が無ければ一覧の先頭を使う（CalendarRepository と同じフォールバック）。
     */
    private suspend fun loadInto(state: DrawerUiState, channelId: String?): DrawerUiState =
        withContext(Dispatchers.IO) {
            val id = channelId ?: channelRepository.getAllChannels().firstOrNull()?.channelID
            val channel = id?.let { channelRepository.getChannel(it) }
            state.copy(
                hasChannel = channel != null,
                channelName = channel?.channelName.orEmpty(),
                myAlias = channel?.myname.orEmpty()
            )
        }
}

data class DrawerItem(
    val title: String,
    val route: String,
    val icon: ImageVector,
    val requiresChannel: Boolean = false
)

@Composable
fun AppDrawer(
    navController: NavController,
    viewModel: DrawerViewModel = hiltViewModel(),
    onClose: () -> Unit = {}
) {
    val uiState by viewModel.uiState.collectAsState()

    ModalDrawerSheet(
        modifier = Modifier.width(280.dp)
    ) {
        // Header
        DrawerHeader(
            channelName = uiState.channelName,
            myAlias = uiState.myAlias,
            onClose = onClose
        )
        
        Divider()
        
        // Navigation items
        LazyColumn(
            modifier = Modifier.fillMaxWidth()
        ) {
            items(getDrawerItems(uiState)) { item ->
                DrawerItemRow(
                    item = item,
                    onClick = {
                        when (item.route) {
                            "home" -> navController.navigate(HomeRoute)
                            "calendar" -> navController.navigate(CalendarRoute())
                            "timestamp" -> navController.navigate(TimestampCodeRoute)
                            "reception" -> navController.navigate(ReceptionRoute())
                            "user" -> navController.navigate(UserRoute)
                            "settings" -> navController.navigate(SettingsRoute)
                            "entryFormEdit" -> navController.navigate(EntryFormListRoute)
                            "tickets" -> navController.navigate(TicketListRoute)
                            "channelCreate" -> navController.navigate(ChannelRoute(id = null))
                            "login" -> navController.navigate(LoginRoute) { launchSingleTop = true }
                            else -> navController.navigate(item.route)
                        }
                        onClose()
                    }
                )
            }
        }
    }
}

@Composable
private fun DrawerHeader(
    channelName: String,
    myAlias: String,
    onClose: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = "Menu",
            style = MaterialTheme.typography.titleLarge,
            fontWeight = FontWeight.Bold
        )
        
        IconButton(onClick = onClose) {
            Icon(
                imageVector = Icons.Default.Close,
                contentDescription = "Close drawer"
            )
        }
    }
    
    if (channelName.isNotBlank()) {
        Text(
            text = "現在のチャネル: $channelName / あなた: $myAlias",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.primary,
            modifier = Modifier.padding(start = 16.dp, end = 16.dp, bottom = 8.dp)
        )
    }
}

@Composable
private fun DrawerItemRow(
    item: DrawerItem,
    onClick: () -> Unit
) {
    NavigationDrawerItem(
        icon = {
            Icon(
                imageVector = item.icon,
                contentDescription = item.title
            )
        },
        label = {
            Text(
                text = item.title,
                fontWeight = FontWeight.Medium
            )
        },
        selected = false,
        onClick = onClick,
        modifier = Modifier.padding(horizontal = 12.dp, vertical = 4.dp)
    )
}

private fun getDrawerItems(uiState: DrawerUiState): List<DrawerItem> {
    val items = mutableListOf<DrawerItem>()

    // Android はサインイン画面を必ず通るため、ドロワーではサインイン有無で出し分けしない。
    // チャネル登録はチャネルを作る導線なので、チャネルが無くても表示する。
    items.add(DrawerItem(title = "ホーム", route = "home", icon = Icons.Default.Home))
    items.add(DrawerItem(title = "ユーザー設定", route = "user", icon = Icons.Default.Person))
    items.add(DrawerItem(title = "システム設定", route = "settings", icon = Icons.Default.Storage))
    items.add(DrawerItem(
        title = "カレンダー",
        route = "calendar",
        icon = Icons.Default.CalendarMonth,
        requiresChannel = true
    ))
    items.add(DrawerItem(
        title = "打刻",
        route = "timestamp",
        icon = Icons.Default.Schedule,
        requiresChannel = true
    ))
    items.add(DrawerItem(
        title = "受付",
        route = "reception",
        icon = Icons.Default.MeetingRoom,
        requiresChannel = true
    ))
    items.add(DrawerItem(
        title = "チャネル登録",
        route = "channelCreate",
        icon = Icons.Default.AddBusiness
    ))
    items.add(DrawerItem(
        title = "フォーム編集",
        route = "entryFormEdit",
        icon = Icons.Default.EditNote,
        requiresChannel = true
    ))
    items.add(DrawerItem(
        title = "チケット",
        route = "tickets",
        icon = Icons.Default.ConfirmationNumber,
        requiresChannel = true
    ))
    items.add(DrawerItem(title = "Sign In", route = "login", icon = Icons.Default.Login))

    // チャネルIDが必要な画面（カレンダー・フォーム編集など）は、チャネルが無ければ一覧に出さない。
    return items.filter { !it.requiresChannel || uiState.hasChannel }
}
