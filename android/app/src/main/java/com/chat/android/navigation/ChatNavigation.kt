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
import com.chat.android.feature.channel.ChannelScreen
import com.chat.android.feature.group.GroupScreen
import com.chat.android.feature.profile.ProfileScreen
import com.chat.android.feature.home.HomeScreen

@Composable
fun ChatNavigation(inviteLink: ProfileRoute? = null) {
    val navController = rememberNavController()

    // 招待URLで起動された場合は、その参加画面を最初の画面にする。
    // LaunchedEffect(Unit) なので再コンポジションでは再度遷移しない。
    LaunchedEffect(Unit) {
        if (inviteLink != null) {
            navController.navigate(inviteLink)
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
