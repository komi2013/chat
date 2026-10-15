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
import com.chat.android.IncomingCallRequest
import com.chat.android.feature.channel.ChannelScreen
import com.chat.android.feature.group.GroupScreen
import com.chat.android.feature.people.PeopleScreen
import com.chat.android.feature.thread.ThreadScreen
import com.chat.android.feature.call.ThreadCallScreen
import com.chat.android.feature.profile.ProfileScreen
import com.chat.android.feature.home.HomeScreen
import com.chat.android.feature.calendar.CalendarScreen
import com.chat.android.navigation.LoginRoute
import com.chat.android.navigation.HomeRoute
import com.chat.android.navigation.CalendarRoute
import com.chat.android.feature.calendar.CalendarEditScreen
import com.chat.android.feature.settings.SettingsScreen
import com.chat.android.feature.reception.ReceptionScreen
import com.chat.android.feature.ticket.TicketsScreen
import com.chat.android.feature.ticket.TicketEditScreen
import kotlinx.serialization.Serializable

/**
 * @param inviteRequest 招待URLで起動・復帰したときの遷移持越し。
 *   LaunchedEffect のキーに使うため、同じURLでも token が変われば再遷移する。
 */
@Composable
fun ChatNavigation(
    inviteRequest: InviteRequest? = null,
    callRequest: IncomingCallRequest? = null,
    isSignedIn: Boolean = false,
    initialChannelID: String? = null
) {
    val navController = rememberNavController()
    // ドロワーのヘッダー表示と「現在のチャネル」追跡で共有する（Activity スコープ）。
    val drawerViewModel: DrawerViewModel = hiltViewModel()

    // 招待URLで起動された場合、その参加画面へ遷移する。
    // キーに inviteRequest を使うことで、起動後にさらに招待リンクを
    // 叩いた場合も onNewIntent -> state 更新 -> 再コンポジションで遷移する。
    LaunchedEffect(inviteRequest) {
        val request = inviteRequest ?: return@LaunchedEffect
        // 未サインインなら csrf が無いため ChannelJoin を呼べない。
        // Vue の Profile.vue と同じくサインイン画面へ送り、TO から再開する。
        if (request.hasSession) {
            navController.navigate(request.route)
        } else {
            navController.navigate(LoginRoute)
        }
    }

    LaunchedEffect(callRequest) {
        val request = callRequest ?: return@LaunchedEffect
        if (isSignedIn) {
            navController.navigate(ThreadCallRoute(request.code)) { launchSingleTop = true }
        } else {
            navController.navigate(LoginRoute) { launchSingleTop = true }
        }
    }

    DrawerLayout(
        navController = navController,
        viewModel = drawerViewModel
    ) { paddingValues ->
        Box(modifier = Modifier.padding(paddingValues)) {
            val startDestination: Any = when {
                !isSignedIn -> LoginRoute
                !initialChannelID.isNullOrBlank() -> ChannelRoute(initialChannelID)
                else -> HomeRoute
            }
            NavHost(navController = navController, startDestination = startDestination) {
                composable<LoginRoute> {
                    SignInScreen(onSignedIn = {
                        val route: Any = callRequest?.let { ThreadCallRoute(it.code) }
                            ?: initialChannelID?.takeIf(String::isNotBlank)?.let { ChannelRoute(it) }
                            ?: HomeRoute
                        navController.navigate(route) {
                            popUpTo(LoginRoute) { inclusive = true }
                            launchSingleTop = true
                        }
                    })
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

                composable<SettingsRoute> {
                    SettingsScreen()
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

                composable<ThreadCallRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<ThreadCallRoute>()
                    ThreadCallScreen(code = route.code, onLeave = { navController.navigateUp() })
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

                composable<ReceptionRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<ReceptionRoute>()
                    ReceptionScreen(
                        navController = navController,
                        id = route.id,
                        code = route.code
                    )
                }

                composable<ReceptionEnterRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<ReceptionEnterRoute>()
                    com.chat.android.feature.reception.ReceptionEnterScreen(
                        id = route.id,
                        passkey = route.passkey,
                        onBack = { navController.navigateUp() }
                    )
                }

                composable<ReceptionMenuRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<ReceptionMenuRoute>()
                    com.chat.android.feature.reception.ReceptionMenuScreen(
                        id = route.id,
                        code = route.code,
                        codeType = route.codeType,
                        onBack = { navController.navigateUp() }
                    )
                }

                composable<ReceptionBookRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<ReceptionBookRoute>()
                    com.chat.android.feature.reception.ReceptionBookScreen(
                        receptionID = route.id,
                        initialDate = route.date,
                        initialMenuID = route.menuID,
                        onBack = { navController.navigateUp() }
                    )
                }

                composable<ReceptionOrderRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<ReceptionOrderRoute>()
                    com.chat.android.feature.reception.ReceptionOrderScreen(
                        id = route.id,
                        code = route.code
                    )
                }

                composable<ReceptionShiftRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<ReceptionShiftRoute>()
                    com.chat.android.feature.reception.ReceptionShiftScreen(
                        receptionID = route.receptionID,
                        onBack = { navController.navigateUp() }
                    )
                }

                composable<ReceptionThreadRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<ReceptionThreadRoute>()
                    com.chat.android.feature.reception.ReceptionThreadScreen(
                        navController = navController,
                        channelID = route.channelID,
                        code = route.code,
                        backID = route.backID
                    )
                }

                composable<TicketListRoute> {
                    TicketsScreen(navController = navController)
                }

                composable<TicketEditRoute> { backStackEntry ->
                    val route = backStackEntry.toRoute<TicketEditRoute>()
                    TicketEditScreen(
                        navController = navController,
                        ticketID = route.ticketID
                    )
                }
            }
        }
    }
}
