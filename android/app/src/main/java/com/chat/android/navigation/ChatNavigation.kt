package com.chat.android.navigation

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.toRoute
import com.chat.android.core.ui.component.DrawerLayout
import com.chat.android.feature.sign.SignInScreen
import com.chat.android.feature.user.UserScreen
import com.chat.android.feature.entryform.EntryFormListScreen
import com.chat.android.feature.entryform.EntryFormEditScreen
import com.chat.android.feature.entryform.EntryFormAnswerScreen
import com.chat.android.InviteRequest
import com.chat.android.feature.channel.ChannelScreen
import com.chat.android.feature.group.GroupScreen
import com.chat.android.feature.profile.ProfileScreen
import com.chat.android.feature.home.HomeScreen

/**
 * @param inviteRequest 招待URLで起動/復帰したときの遷移指示。
 *   LaunchedEffect のキーに使うため、同じURLでも token が変われば再遷移する。
 */
@Composable
fun ChatNavigation(inviteRequest: InviteRequest? = null) {
    val navController = rememberNavController()

    // 招待URLで起動された場合、その参加画面へ遷移する。
    // キーに inviteRequest を使うことで、起動後にさらに招待リンクを
    // 叩いた場合も（onNewIntent → state 更新 → 再コンポジション）遷移する。
    LaunchedEffect(inviteRequest) {
        val request = inviteRequest ?: return@LaunchedEffect
        // 未サインインなら csrf が無いため ChannelJoin を呼べない。
        // Vue の Profile.vue と同じくサインイン画面へ送り、TO から再開する。
        if (request.hasSession) {
            navController.navigate(request.route)
        } else {
            navController.navigate("login")
        }
    }

    DrawerLayout(
        navController = navController
    ) { paddingValues ->
        Box(modifier = Modifier.padding(paddingValues)) {
            NavHost(navController = navController, startDestination = "login") {
                composable("login") {
                    SignInScreen(navController = navController)
                }
                
                composable<HomeRoute> {
                    HomeScreen(navController = navController)
                }
                
                composable<UserRoute> {
                    UserScreen(navController = navController)
                }

                composable<EntryFormListRoute> {
                    EntryFormListScreen(navController = navController)
                }

                composable<EntryFormEditRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<EntryFormEditRoute>()
                    EntryFormEditScreen(
                        navController = navController,
                        id = route.id,
                        formJson = route.formJson
                    )
                }

                composable<EntryFormAnswerRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<EntryFormAnswerRoute>()
                    EntryFormAnswerScreen(
                        navController = navController,
                        id = route.id
                    )
                }

                composable<ChannelRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<ChannelRoute>()
                    ChannelScreen(
                        navController = navController,
                        id = route.id
                    )
                }

                composable<ProfileRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<ProfileRoute>()
                    ProfileScreen(
                        navController = navController,
                        id = route.id,
                        code = route.code
                    )
                }

                composable<GroupRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<GroupRoute>()
                    GroupScreen(
                        navController = navController,
                        id = route.id
                    )
                }
            }
        }
    }
}
