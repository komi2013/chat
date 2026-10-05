package com.chat.android.navigation

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.toRoute
import androidx.hilt.navigation.compose.hiltViewModel
import com.chat.android.core.ui.component.DrawerLayout
import com.chat.android.core.ui.component.DrawerViewModel
import com.chat.android.feature.sign.SignInScreen
import com.chat.android.feature.user.UserScreen
import com.chat.android.feature.entryform.EntryFormListScreen
import com.chat.android.feature.entryform.EntryFormEditScreen
import com.chat.android.feature.entryform.EntryFormAnswerScreen
import com.chat.android.InviteRequest
import com.chat.android.feature.channel.ChannelScreen
import com.chat.android.feature.group.GroupScreen
import com.chat.android.feature.people.PeopleScreen
import com.chat.android.feature.thread.ThreadScreen
import com.chat.android.feature.profile.ProfileScreen
import com.chat.android.feature.home.HomeScreen
import com.chat.android.feature.calendar.CalendarScreen
import com.chat.android.feature.calendar.CalendarEditScreen

/**
 * @param inviteRequest 諡帛ｾ・RL縺ｧ襍ｷ蜍・蠕ｩ蟶ｰ縺励◆縺ｨ縺阪・驕ｷ遘ｻ謖・､ｺ縲・
 *   LaunchedEffect 縺ｮ繧ｭ繝ｼ縺ｫ菴ｿ縺・◆繧√∝酔縺篭RL縺ｧ繧・token 縺悟､峨ｏ繧後・蜀埼・遘ｻ縺吶ｋ縲・
 */
@Composable
fun ChatNavigation(inviteRequest: InviteRequest? = null) {
    val navController = rememberNavController()
    // 繝峨Ο繝ｯ繝ｼ縺ｮ繝倥ャ繝繝ｼ陦ｨ遉ｺ縺ｨ縲檎樟蝨ｨ縺ｮ繝√Ε繝阪Ν縲崎ｿｽ霍｡縺ｧ蜈ｱ譛峨☆繧具ｼ・ctivity 繧ｹ繧ｳ繝ｼ繝暦ｼ峨・
    val drawerViewModel: DrawerViewModel = hiltViewModel()

    // 諡帛ｾ・RL縺ｧ襍ｷ蜍輔＆繧後◆蝣ｴ蜷医√◎縺ｮ蜿ょ刈逕ｻ髱｢縺ｸ驕ｷ遘ｻ縺吶ｋ縲・
    // 繧ｭ繝ｼ縺ｫ inviteRequest 繧剃ｽｿ縺・％縺ｨ縺ｧ縲∬ｵｷ蜍募ｾ後↓縺輔ｉ縺ｫ諡帛ｾ・Μ繝ｳ繧ｯ繧・
    // 蜿ｩ縺・◆蝣ｴ蜷医ｂ・・nNewIntent 竊・state 譖ｴ譁ｰ 竊・蜀阪さ繝ｳ繝昴ず繧ｷ繝ｧ繝ｳ・蛾・遘ｻ縺吶ｋ縲・
    LaunchedEffect(inviteRequest) {
        val request = inviteRequest ?: return@LaunchedEffect
        // 譛ｪ繧ｵ繧､繝ｳ繧､繝ｳ縺ｪ繧・csrf 縺檎┌縺・◆繧・ChannelJoin 繧貞他縺ｹ縺ｪ縺・・
        // Vue 縺ｮ Profile.vue 縺ｨ蜷後§縺上し繧､繝ｳ繧､繝ｳ逕ｻ髱｢縺ｸ騾√ｊ縲ゝO 縺九ｉ蜀埼幕縺吶ｋ縲・
        if (request.hasSession) {
            navController.navigate(request.route)
        } else {
            navController.navigate("login")
        }
    }

    DrawerLayout(
        navController = navController,
        viewModel = drawerViewModel
    ) { paddingValues ->
        Box(modifier = Modifier.padding(paddingValues)) {
            NavHost(navController = navController, startDestination = "login") {
                composable("login") {
                    SignInScreen(navController = navController)
                }
                
                composable<HomeRoute> {
                    HomeScreen(navController = navController)
                }

                composable<CalendarRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<CalendarRoute>()
                    CalendarScreen(navController = navController, date = route.date)
                }

                composable<CalendarEditRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<CalendarEditRoute>()
                    CalendarEditScreen(navController = navController, id = route.id, dateTime = route.dateTime)
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
                    LaunchedEffect(route.id) {
                        route.id?.let { drawerViewModel.onChannelOpened(it) }
                    }
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

                composable<PeopleRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<PeopleRoute>()
                    PeopleScreen(
                        navController = navController,
                        id = route.id,
                        name = route.name
                    )
                }

                composable<ThreadRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<ThreadRoute>()
                    LaunchedEffect(route.channelID) {
                        route.channelID?.let { drawerViewModel.onChannelOpened(it) }
                    }
                    ThreadScreen(
                        navController = navController,
                        channelID = route.channelID,
                        parentID = route.parentID,
                        backID = route.backID
                    )
                }

                composable<TimestampRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<TimestampRoute>()
                    com.chat.android.feature.timestamp.TimestampScreen(
                        navController = navController,
                        adminName = route.adminName,
                        code = route.code
                    )
                }

                composable<TimestampCodeRoute> {
                    com.chat.android.feature.timestamp.TimestampCodeScreen(
                        navController = navController
                    )
                }

                composable<TimestampReportRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<TimestampReportRoute>()
                    com.chat.android.feature.timestamp.TimestampReportScreen(
                        navController = navController,
                        admin = route.admin,
                        month = route.month,
                        stamper = route.stamper
                    )
                }
            }
        }
    }
}

