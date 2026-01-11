package controller

import (
  "chat/collection"
  "chat/common"
  "context"
  "encoding/json"
  // "log"
  // "math"
  "net/http"
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

  // SOL残高を取得
  var solBalance *float64
  var userSolanaWalletAddress string
  if user.SolanaWalletAddress != "" {
    userSolanaWalletAddress = user.SolanaWalletAddress
    
    // Solana RPC接続
    client := rpc.New(rpc.MainNetBeta_RPC)
    
    // ウォレットアドレスをパース
    walletPubkey, err := solana.PublicKeyFromBase58(user.SolanaWalletAddress)
    if err == nil {
      // SOL残高を取得
      balance, err := client.GetBalance(ctx, walletPubkey, rpc.CommitmentConfirmed)
      if err == nil {
        // 残高をSOL単位に変換（1 SOL = 1,000,000,000 lamports）
        balanceValue := float64(balance) / float64(1_000_000_000)
        solBalance = &balanceValue
      } else {
        // エラーの場合（残高0）
        zeroBalance := 0.0
        solBalance = &zeroBalance
      }
    }
  }

  responseData := struct {
    Csrf         string       `json:"csrf"`
    PushContents []string     `json:"pushContents"`
    User       collection.UserResponse  `json:"user"`
    Nicknames   []collection.NicknameResponse `json:"nicknames"`
    JpycBalance            *float64   `json:"jpycBalance,omitempty"` // 互換性のため保持（SOL残高を返す）
    SolanaWalletAddress    string     `json:"solanaWalletAddress,omitempty"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    User         : user,
    Nicknames: nicknames,
    JpycBalance: solBalance, // SOL残高を返す（互換性のためフィールド名は変更しない）
    SolanaWalletAddress: userSolanaWalletAddress,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

