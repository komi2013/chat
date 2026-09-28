package com.chat.android.feature.sign

import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.navigation.UserRoute
import com.google.android.gms.auth.api.signin.GoogleSignIn
import com.google.android.gms.common.api.ApiException

@Composable
fun SignInScreen(
    navController: NavController,
    viewModel: SignInViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    val launcher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.StartActivityForResult()
    ) { result ->
        val task = GoogleSignIn.getSignedInAccountFromIntent(result.data)
        try {
            val account = task.getResult(ApiException::class.java)
            viewModel.handleGoogleSignInResult(account.idToken)
        } catch (e: ApiException) {
            viewModel.handleGoogleSignInResult(null, "Google Sign-In failed: ${e.statusCode}")
        }
    }

    LaunchedEffect(uiState) {
        if (uiState is SignInUiState.Success) {
            navController.navigate(UserRoute) {
                popUpTo("login") { inclusive = true }
            }
        }
    }

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
            modifier = Modifier.padding(bottom = 32.dp)
        )

        if (uiState is SignInUiState.Error) {
            Text(
                text = (uiState as SignInUiState.Error).message,
                color = MaterialTheme.colorScheme.error,
                modifier = Modifier.padding(bottom = 16.dp)
            )
        }

        if (uiState is SignInUiState.Loading) {
            CircularProgressIndicator(
                modifier = Modifier.padding(bottom = 16.dp)
            )
        }
        
        Button(
            onClick = {
                try {
                    android.util.Log.d("SignInScreen", "Launching Google Sign-In intent")
                    launcher.launch(viewModel.getGoogleSignInClient().signInIntent)
                } catch (e: Exception) {
                    android.util.Log.e("SignInScreen", "Crash during intent launch", e)
                    viewModel.handleGoogleSignInResult(null, "Failed to start Google Sign-In: ${e.message}")
                }
            },
            modifier = Modifier.fillMaxWidth(),
            enabled = uiState !is SignInUiState.Loading
        ) {
            Text("Sign in with Google")
        }
    }
}
