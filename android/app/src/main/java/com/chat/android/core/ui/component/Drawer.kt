package com.chat.android.core.ui.component

import androidx.compose.foundation.background
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
import com.chat.android.core.network.ChannelDetail
import com.chat.android.core.network.NicknameResponse
import com.chat.android.feature.channel.ChannelRepository
import com.chat.android.feature.channel.DbChannel
import com.chat.android.navigation.UserRoute
import com.chat.android.navigation.EntryFormListRoute
import com.chat.android.navigation.EntryFormEditRoute
import com.chat.android.navigation.ChannelRoute
import com.chat.android.navigation.HomeRoute
import com.chat.android.navigation.CalendarRoute
import javax.inject.Inject
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

data class DrawerUiState(
    val isSignedIn: Boolean = false,
    val hasChannel: Boolean = false,
    val isGuest: Boolean = false,
    val channelName: String = "",
    val myAlias: String = "",
    val myImg: String = ""
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
        viewModelScope.launch {
            combine(userRepository.signInState, userRepository.currentChannelId) { isSignedIn, channelId ->
                isSignedIn to channelId
            }.collect { (isSignedIn, channelId) ->
                _uiState.value = loadInto(DrawerUiState(isSignedIn = isSignedIn), channelId)
            }
        }
        // エイリアス切替・チャネル編集はローカルDBを書き換えるので、ヘッダー表示も更新する。
        viewModelScope.launch {
            channelRepository.dbUpdateFlow.collect {
                _uiState.value = loadInto(_uiState.value, userRepository.currentChannelId.value)
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

    /** 画面遷移やサインイン状態の変化時に呼ぶ。 */
    fun refresh(channel: ChannelDetail? = null) {
        viewModelScope.launch {
            val loaded = loadInto(_uiState.value, userRepository.currentChannelId.value)
            _uiState.value = if (channel != null) {
                loaded.copy(
                    hasChannel = true,
                    channelName = channel.name,
                    myAlias = channel.myname.orEmpty()
                )
            } else loaded
        }
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
                myAlias = channel?.myname.orEmpty(),
                myImg = channel?.myimg.orEmpty()
            )
        }
}

data class DrawerItem(
    val title: String,
    val route: String,
    val icon: ImageVector,
    val requiresAuth: Boolean = false,
    val requiresChannel: Boolean = false,
    val requiresNonGuest: Boolean = false
)

@Composable
fun AppDrawer(
    navController: NavController,
    channel: ChannelDetail? = null,
    aliases: List<NicknameResponse> = emptyList(),
    viewModel: DrawerViewModel = hiltViewModel(),
    onClose: () -> Unit = {}
) {
    val uiState by viewModel.uiState.collectAsState()
    
    LaunchedEffect(channel, aliases) {
        viewModel.refresh(channel)
    }

    ModalDrawerSheet(
        modifier = Modifier.width(280.dp)
    ) {
        // Header
        DrawerHeader(
            isSignedIn = uiState.isSignedIn,
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
                            "user" -> navController.navigate(UserRoute)
                            "entryFormEdit" -> navController.navigate(EntryFormListRoute)
                            "channelCreate" -> navController.navigate(ChannelRoute(id = null))
                            else -> navController.navigate(item.route)
                        }
                        onClose()
                    }

                )
            }
        }
        
        // Advertisement section (like in Vue)
        if (uiState.isSignedIn) {
            Divider()
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "Advertisement",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}

@Composable
private fun DrawerHeader(
    isSignedIn: Boolean,
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
    
    when {
        channelName.isNotBlank() -> {
            Text(
                text = "現在のチャネル: $channelName / あなた: $myAlias",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.primary,
                modifier = Modifier.padding(start = 16.dp, end = 16.dp, bottom = 8.dp)
            )
        }
        isSignedIn -> {
            Text(
                text = "Signed In",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.primary,
                modifier = Modifier.padding(start = 16.dp, bottom = 8.dp)
            )
        }
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
    
    // Home item
    items.add(DrawerItem(
        title = "ホーム",
        route = "home",
        icon = Icons.Default.Home
    ))

    // Authentication-dependent items
    if (uiState.isSignedIn) {
        items.add(DrawerItem(
            title = "ユーザー設定", 
            route = "user",
            icon = Icons.Default.Person,
            requiresAuth = true
        ))
    }

    if (uiState.isSignedIn) {
        items.add(DrawerItem(
            title = "カレンダー",
            route = "calendar",
            icon = Icons.Default.CalendarMonth,
            requiresAuth = true,
            requiresChannel = true
        ))
    }
    
    // チャネル依存の画面。チャネル登録はチャネルを作る導線なので、チャネルが無くても表示する。
    if (uiState.isSignedIn) {
        items.add(DrawerItem(
            title = "チャネル登録",
            route = "channelCreate",
            icon = Icons.Default.AddBusiness,
            requiresAuth = true
        ))

        items.add(DrawerItem(
            title = "フォーム編集",
            route = "entryFormEdit",
            icon = Icons.Default.EditNote,
            requiresAuth = true,
            requiresChannel = true
        ))
    }
    
    if (!uiState.isSignedIn) {
        items.add(DrawerItem(
            title = "サインイン",
            route = "login",
            icon = Icons.Default.Login
        ))
    }
    
    // チャネルIDが必要な画面（カレンダー・フォーム編集など）は、チャネルが無ければ一覧に出さない。
    return items.filter { !it.requiresChannel || uiState.hasChannel }
}
