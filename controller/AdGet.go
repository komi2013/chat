package controller

import (
  "chat/collection"
  "chat/common"
  "context"
  "encoding/json"
  "fmt"
  // "log"
  // "math"
  "net/http"
  "strconv"
  "time"

  "github.com/gagliardetto/solana-go"
  "github.com/gagliardetto/solana-go/rpc"
  "go.mongodb.org/mongo-driver/bson"
)

func AdGet(w http.ResponseWriter, r *http.Request) {

  ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
  defer cancel()

  session, err := common.SessionCheckTake(w, r, r.FormValue("csrf"))
  if err != nil {
    common.WriteResponseWithoutSession(w, r.FormValue("csrf"), err.Error()+";SessionCheckTake", http.StatusOK)
    return
  }

  coll := common.DB.AdDB.Collection("ad")
  filter := bson.M{
    "userID": session.UserID,
  }
	var ads []collection.AdStruct
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";coll.Find", http.StatusOK)
    return
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &ads); err != nil {
    common.WriteResponseWithSession(w, session, err.Error()+";cursor.All", http.StatusOK)
    return
	}

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	for i := range ads {
		// AdStart, AdEnd を JST に変換
		ads[i].AdStart = ads[i].AdStart.In(jst)
		ads[i].AdEnd = ads[i].AdEnd.In(jst)
		// UpdatedAt や CreatedAt など他の時刻フィールドがある場合も同様に変換します
		ads[i].UpdatedAt = ads[i].UpdatedAt.In(jst)
		// 請求・支払い関連の日付も念のため変換 (ゼロ値でない場合)
		if !ads[i].InvoicedAt.IsZero() {
			ads[i].InvoicedAt = ads[i].InvoicedAt.In(jst)
		}
		if !ads[i].PaidAt.IsZero() {
			ads[i].PaidAt = ads[i].PaidAt.In(jst)
		}
	}

  // システムFee Payerの公開鍵を取得
  cfg := common.LoadConfig()
  var systemFeePayerPublicKey string
  if cfg.SolanaFeePayerPrivateKey != "" {
    pubkey, err := common.GetSystemFeePayerPublicKey(cfg.SolanaFeePayerPrivateKey)
    if err == nil {
      systemFeePayerPublicKey = pubkey.String()
    } else {
      // 仮の値（コンパイルエラー回避）
      systemFeePayerPublicKey = ""
    }
  }

  // Solanaシステムウォレットアドレスを取得（Fee Payerの公開鍵を使用）
  var systemSolanaWalletAddress string
  if cfg.SolanaFeePayerPrivateKey != "" {
    pubkey, err := common.GetSystemFeePayerPublicKey(cfg.SolanaFeePayerPrivateKey)
    if err == nil {
      systemSolanaWalletAddress = pubkey.String()
    }
  }

  // Solanaアドレス履歴確認URLを生成（Solscanを使用）
  var jpycCheckURL string
  if systemSolanaWalletAddress != "" {
    jpycCheckURL = fmt.Sprintf("https://solscan.io/account/%s", systemSolanaWalletAddress)
  }

  // ユーザー情報を取得してJPYC残高を取得
  var jpycBalance *float64
  var userSolanaWalletAddress string
  collUser := common.DB.UserDB.Collection("user")
  filterUser := bson.M{"_id": session.UserID}
  var user collection.UserStruct
  err = collUser.FindOne(ctx, filterUser).Decode(&user)
  if err == nil && user.SolanaWalletAddress != "" {
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
    Ads       []collection.AdStruct  `json:"ads"`
    SystemWalletAddress   string       `json:"systemWalletAddress"`
    SystemFeeWalletAddress string      `json:"systemFeeWalletAddress,omitempty"`
    SystemFeePayerPublicKey string     `json:"systemFeePayerPublicKey,omitempty"`
    JpycMintAddress        string     `json:"jpycMintAddress,omitempty"`
    JpycCheckURL           string     `json:"jpycCheckURL,omitempty"`
    JpycBalance            *float64   `json:"jpycBalance,omitempty"`
    SolanaWalletAddress    string     `json:"solanaWalletAddress,omitempty"`
  }{
    Csrf:         session.Csrf,
    PushContents: session.PushContents,
    Ads:          ads,
    SystemWalletAddress: systemSolanaWalletAddress, // Solanaアドレスを使用
    SystemFeeWalletAddress: systemSolanaWalletAddress, // デフォルトはシステムウォレット
    SystemFeePayerPublicKey: systemFeePayerPublicKey,
    JpycMintAddress: common.JpycSolanaMint,
    JpycCheckURL: jpycCheckURL,
    JpycBalance: jpycBalance,
    SolanaWalletAddress: userSolanaWalletAddress,
  }

  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(responseData)
}

