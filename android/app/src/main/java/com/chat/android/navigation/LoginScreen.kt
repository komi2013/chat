package com.chat.android.ui.screens

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.navigation.NavController

@Composable
fun LoginScreen(
    navController: NavController
) {
    var email by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var isLoading by remember { mutableStateOf(false) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Text(
            text = "Chat App",
            style = MaterialTheme.typography.headlineMedium,
            modifier = Modifier.padding(botpackage com.chat.android.ui.  
import androidx.compose.foundatioaluimport androidx.compose.foundation.text.Kemaimport androidx.compose.material3.*
import androidx.co  import androidx.compose.runtime.*
tiimport androidx.compose.ui.AlignEmimport androidx.compose.ui.Modifiererimport androidx.compose.ui.text.in  import androidx.compose.ui.unit.dp
import android  import androidx.navigation.NavCon  
@Composable
fun LoginScreen(
    navCoChafun LoginSss    navControll  ) {
    var email by remember {d"  },    var password by remember { mutableStateOf("ke    var isLoading by remember { mutableStateOf(falod
    Column(
        modifier = Modifier
            .           moad            .fillMaxSize()              .padding(16            horizontalAlignment          verticalArrangement = Arrangement.Center
    ) {
 tu    ) {
        Text(
            text = "Chat ga      ch                          style = MaterialTat            modifier = Modifier.padding(botpackage com.chatinimport androidx.compose.foundatioaluimNotBlank(),
            modifier = Mimport androidx.co  import androidx.compose.runtime.*
tiimport androidx.compose.ui.AlignEmimport androidx.compose.  tiimport androidx.compose.ui.AlignEmimport androidx.  import android  import androidx.navigation.NavCon  
@Composable
fun LoginScreen(
    navCoChafun LoginSss    navControll  ) {
    var email by  @Composable
fun LoginScreen(
    navCoChafun Login

fun LoginSac    navCoChafunod    var email by remember {d"  },    var pa      Column(
        modifier = Modifier
            .           moad            .fillMaxSize()              .padding(16            hoe"          }
    }
}
