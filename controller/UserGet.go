package controller

import (
  "chat/collection"
  "chat/common"
  "context"
  "encoding/json"
  // "log"
  // "math"
  "net/http"
  "strconv"
  // "time"

  "github.com/gagliardetto/solana-go"
  "github.com/gagliardetto/solana-go/rpc"
  "go.mongodb.org/mongo-driver/bson"
  // "go.mongodb.org/mongo-driver/mongo"
  // "go.mongodb.org/mongo-driver/mongo/options"
)

func UserGet(w http.ResponseWriter, r *http.Request) {
  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
  	common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";Session Check", http.StatusOK)
    return
  }

  ctx := context.TODO()
  collUser := common.DB.UserDB.Collection("user")
  filterUser := bson.M{"_id": session.UserID}
  var user collection.UserResponse
  err = collUser.FindOne(ctx, filterUser).Decode(&user)
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";collUser.Find", http.StatusOK)
    return
  }
  collNickname := common.DB.NicknameDB.Collection("nickname")
  filterNickname := bson.M{"userID": session.UserID}
  cursor, err := collNickname.Find(ctx, filterNickname)
  if err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";collNickname.Find", http.StatusOK)
    return
  }
  var nicknames []collection.NicknameResponse
  if err = cursor.All(ctx, &nicknames); err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";cursor.All", http.StatusOK)
    return
  }

  // JPYC残高を取得
  var jpycBalance *float64
  var userSolanaWalletAddress string
  if user.SolanaWalletAddress != "" {
    userSolanaWalletAddress = user.SolanaWalletAddress
    
    // Solana RPC接続
    client := rpc.New(rpc.MainNetBeta_RPC)
    
    // ウォレットアドレスをパース
    walletPubkey, err := solana.PublicKeyFromBase58(user.SolanaWalletAddress)
    if err == nil {
      // JPYC Mintアドレス
      jpycMint, err := solana.PublicKeyFromBase58(common.JpycSolanaMint)
      if err == nil {
        // トークンアカウントアドレスを取得
        tokenAccount, err := common.FindAssociatedTokenAddress(
          walletPubkey,
          jpycMint,
        )
        if err == nil {
          // 残高を取得
          balance, err := client.GetTokenAccountBalance(ctx, tokenAccount, rpc.CommitmentConfirmed)
          if err == nil && balance.Value != nil {
            // 残高をJPYC単位に変換
            decimals := balance.Value.Decimals
            if decimals == 0 {
              decimals = 6 // デフォルト値
            }
            amountStr := balance.Value.Amount
            amount, err := strconv.ParseUint(amountStr, 10, 64)
            if err == nil {
              balanceValue := float64(amount) / float64(1e6) // decimals=6を想定
              jpycBalance = &balanceValue
            } else {
              // パースエラーの場合（残高0）
              zeroBalance := 0.0
              jpycBalance = &zeroBalance
            }
          } else {
            // トークンアカウントが存在しない場合（残高0）
            zeroBalance := 0.0
            jpycBalance = &zeroBalance
          }
        }
      }
    }
  }

  responseData := struct {
    Csrf         string       `json:"csrf"`
    PushContents []string     `json:"pushContents"`
    User       collection.UserResponse  `json:"user"`
    Nicknames   []collection.NicknameResponse `json:"nicknames"`
    JpycBalance            *float64   `json:"jpycBalance,omitempty"`
    SolanaWalletAddress    string     `json:"solanaWalletAddress,omitempty"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    User         : user,
    Nicknames: nicknames,
    JpycBalance: jpycBalance,
    SolanaWalletAddress: userSolanaWalletAddress,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

