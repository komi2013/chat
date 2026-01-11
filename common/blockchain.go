package common

import (
    "context"
    "crypto/ecdsa"
    // "crypto/sha256"
    "encoding/base64"
    "encoding/binary"
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "math/big"
    "net/http"
    "strings"

    "github.com/ethereum/go-ethereum/crypto"
    hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
    bin "github.com/gagliardetto/binary"
    "github.com/gagliardetto/solana-go"
    "github.com/gagliardetto/solana-go/programs/token"
    "github.com/gagliardetto/solana-go/rpc"
)

//
// ====== 基本設定 ======
//

// マスターシード（必ず環境変数に移すこと）
const masterSeed = "legal winner thank year wave sausage worth useful legal winner thank yellow"

// チェーン（Polygon / Ethereum）定義
const (
    ChainPolygon  = "polygon"
    ChainEthereum = "ethereum"
)

// JPYC コントラクト（EVM 共通）
const JpycContract = "0xE7C3D8C9a439feDe00D2600032D5dB0Be71C3c29"

// API エンドポイント
var chainAPI = map[string]string{
    ChainPolygon:  "https://polygon.blockscout.com/api",
    ChainEthereum: "https://api.etherscan.io/api",
}

//
// ====== アドレス生成 ======
//

const PolygonAPI = "https://polygon.blockscout.com"
const SystemWalletAddress = "0x0dc24f370ceb37895112d3bc53631e2aef9fb1a4"

// Solana関連の定数
// JPYC SPL Token Mint Address (Solana Mainnet)
// 注意: 実際のJPYC Mintアドレスに置き換えてください
const JpycSolanaMint = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" // これはUSDCの例。実際のJPYC Mintアドレスに変更が必要
const SystemFeeWalletAddress = "" // システム利用料を受け取るウォレットアドレス（環境変数から設定）
const SystemFeeAmount = uint64(1_000_000) // 1 JPYC (decimals=6の場合) = 1,000,000

// GetSystemFeePayerPublicKey システムFee Payerの公開鍵を取得
func GetSystemFeePayerPublicKey(privateKeyBase58 string) (solana.PublicKey, error) {
    if privateKeyBase58 == "" {
        // 仮の値（コンパイルエラー回避）
        return solana.PublicKey{}, fmt.Errorf("private key is empty")
    }
    privateKey, err := solana.PrivateKeyFromBase58(privateKeyBase58)
    if err != nil {
        return solana.PublicKey{}, err
    }
    return privateKey.PublicKey(), nil
}

// FindAssociatedTokenAddress 関連トークンアカウントアドレスを計算する
// Associated Token Program ID: ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL
// Seeds: [owner, Token Program ID, mint]
func FindAssociatedTokenAddress(owner solana.PublicKey, mint solana.PublicKey) (solana.PublicKey, error) {
	// Associated Token Program ID（正しいアドレス）
	ataProgramID := solana.MustPublicKeyFromBase58("ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL")
	
	// Seeds: [owner, Token Program ID, mint]
	seeds := [][]byte{
		owner.Bytes(),
		token.ProgramID.Bytes(),
		mint.Bytes(),
	}
	
	// solana.FindProgramAddressを使用してPDAを計算
	// この関数は、Ed25519曲線上にない点を見つけるまでbump seedを試行します
	address, _, err := solana.FindProgramAddress(seeds, ataProgramID)
	if err != nil {
		return solana.PublicKey{}, fmt.Errorf("failed to find associated token address: %w", err)
	}
	
	return address, nil
}

// HD Wallet 生成（共通）
func GenerateInvoiceAddress(invoiceID string) (string, error) {
    wallet, err := hdwallet.NewFromMnemonic(masterSeed)
    if err != nil {
        return "", err
    }

    // m/44'/60'/0'/0/{index}
    path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%s", invoiceID))

    account, err := wallet.Derive(path, false)
    if err != nil {
        return "", err
    }

    return account.Address.Hex(), nil
}

// 出金したい場合だけ秘密鍵が必要
func GetPrivateKey(invoiceID string) (*ecdsa.PrivateKey, error) {
    wallet, err := hdwallet.NewFromMnemonic(masterSeed)
    if err != nil {
        return nil, err
    }

    path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%s", invoiceID))
    account, err := wallet.Derive(path, false)
    if err != nil {
        return nil, err
    }

    return wallet.PrivateKey(account)
}

func GetPublicKeyHex(priv *ecdsa.PrivateKey) string {
    return crypto.PubkeyToAddress(priv.PublicKey).Hex()
}

//
// ====== 入金確認（共通 API） ======
//

// Blockscout/Etherscan 共通レスポンス
type TokenTxResponse struct {
    Status  string    `json:"status"`
    Message string    `json:"message"`
    Result  []TokenTx `json:"result"`
}

type TokenTx struct {
    Hash        string `json:"hash"`
    From        string `json:"from"`
    To          string `json:"to"`
    Value       string `json:"value"`
    TimeStamp   string `json:"timeStamp"`
    BlockNumber string `json:"blockNumber"`
}

//
// JPYC 入金チェック（チェーンを指定）
//
func CheckJPYCReceived(chain, invoiceAddress string, expectAmount float64) (*TokenTx, error) {

    api := chainAPI[chain]
    if api == "" {
        return nil, fmt.Errorf("unsupported chain: %s", chain)
    }

    url := fmt.Sprintf(
        "%s?module=account&action=tokentx&address=%s&contractaddress=%s&page=1&offset=20&sort=desc",
        api,
        invoiceAddress,
        JpycContract,
    )

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)

    var data TokenTxResponse
    if err := json.Unmarshal(body, &data); err != nil {
        return nil, err
    }

    if data.Status != "1" {
        return nil, fmt.Errorf("API error: %s", data.Message)
    }

    for _, tx := range data.Result {
        if !strings.EqualFold(tx.To, invoiceAddress) {
            continue
        }

        amt := WeiToJPYC(tx.Value)
        if amt >= expectAmount {
            return &tx, nil
        }
    }

    return nil, nil
}

//
// Wei → JPYC（小数）
//
func WeiToJPYC(v string) float64 {
    num := new(big.Float)
    num.SetString(v)

    base := new(big.Float).SetFloat64(1e18)

    out := new(big.Float).Quo(num, base)
    result, _ := out.Float64()
    return result
}

//
// ====== Solana/JPY 為替レート取得 ======
//

// SolToJpyRateResponse CoinGecko API レスポンス構造体
type SolToJpyRateResponse struct {
    Solana struct {
        JPY float64 `json:"jpy"`
    } `json:"solana"`
}

// GetSolToJpyRate CoinGecko API から SOL/JPY 為替レートを取得する
func GetSolToJpyRate() (float64, error) {
    url := "https://api.coingecko.com/api/v3/simple/price?ids=solana&vs_currencies=jpy"
    
    resp, err := http.Get(url)
    if err != nil {
        return 0, fmt.Errorf("failed to fetch exchange rate: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return 0, fmt.Errorf("API returned status code: %d", resp.StatusCode)
    }
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return 0, fmt.Errorf("failed to read response body: %w", err)
    }
    
    var data SolToJpyRateResponse
    if err := json.Unmarshal(body, &data); err != nil {
        return 0, fmt.Errorf("failed to parse JSON: %w", err)
    }
    
    if data.Solana.JPY == 0 {
        return 0, errors.New("exchange rate is zero or not found")
    }
    
    return data.Solana.JPY, nil
}

// YenToSol 円からSOLに変換
func YenToSol(yen float64, rate float64) float64 {
    if rate == 0 {
        return 0
    }
    return yen / rate
}

// SolToYen SOLから円に変換
func SolToYen(sol float64, rate float64) float64 {
    return sol * rate
}

//
// ====== Solana トランザクション検証 ======
//

// VerifySolanaTransaction Solana トランザクションを検証する
func VerifySolanaTransaction(signatureStr string, merchantWalletAddr string, expectedLamportsAmount uint64) error {
    client := rpc.New(rpc.MainNetBeta_RPC)

    sig, err := solana.SignatureFromBase58(signatureStr)
    if err != nil {
        return err
    }

    merchantWallet, err := solana.PublicKeyFromBase58(merchantWalletAddr)
    if err != nil {
        return fmt.Errorf("invalid merchant wallet address: %w", err)
    }

    out, err := client.GetTransaction(
        context.Background(),
        sig,
        &rpc.GetTransactionOpts{
            Commitment: rpc.CommitmentFinalized,
            Encoding:   solana.EncodingJSON,
        },
    )
    if err != nil || out == nil {
        return errors.New("transaction not found")
    }

    meta := out.Meta
    if meta == nil || meta.Err != nil {
        return errors.New("transaction failed")
    }

    // トランザクションを取得
    tx, err := out.Transaction.GetTransaction()
    if err != nil {
        return fmt.Errorf("failed to get transaction: %w", err)
    }

    // 送金内容を検証
    msg := tx.Message
    found := false

    for _, inst := range msg.Instructions {
        // ProgramIDIndex を使ってプログラムIDを取得
        if int(inst.ProgramIDIndex) >= len(msg.AccountKeys) {
            continue
        }
        programID := msg.AccountKeys[inst.ProgramIDIndex]
        
        if programID.Equals(solana.SystemProgramID) {
            accounts := inst.Accounts
            if len(accounts) < 2 {
                continue
            }

            // Accounts はインデックスの配列
            if int(accounts[1]) >= len(msg.AccountKeys) {
                continue
            }
            to := msg.AccountKeys[accounts[1]]
            if to.Equals(merchantWallet) {
                found = true
            }
        }
    }

    if !found {
        return errors.New("merchant wallet mismatch")
    }

    // 金額検証（balance差分）
    pre := meta.PreBalances
    post := meta.PostBalances

    var received uint64
    for i, k := range msg.AccountKeys {
        if k.Equals(merchantWallet) {
            received = post[i] - pre[i]
        }
    }

    if received != expectedLamportsAmount {
        return fmt.Errorf("amount mismatch: expected %d, received %d", expectedLamportsAmount, received)
    }

    return nil
}

//
// ====== ガスレス決済（Fee Relayer）関連 ======
//

// PaymentExecuteRequest 決済実行リクエスト
type PaymentExecuteRequest struct {
    TransactionBase64 string `json:"transactionBase64"` // ユーザーが部分署名したトランザクション（base64）
    AdID              string `json:"adID"`              // 広告ID
}

// PaymentExecuteResponse 決済実行レスポンス
type PaymentExecuteResponse struct {
    Signature string `json:"signature"` // トランザクション署名
    Error    string `json:"error,omitempty"`
}

// ExecutePaymentWithFeeRelayer ユーザー署名済みトランザクションを受け取り、Fee Payerとして署名を追加して送信
func ExecutePaymentWithFeeRelayer(
    ctx context.Context,
    transactionBase64 string,
    feePayerPrivateKeyBase58 string,
    systemFeeWallet solana.PublicKey,
    expectedSystemFeeAmount uint64,
) (string, error) {
    // 1. トランザクションをデシリアライズ
    txBytes, err := base64.StdEncoding.DecodeString(transactionBase64)
    if err != nil {
        return "", fmt.Errorf("failed to decode transaction: %w", err)
    }

    var tx solana.Transaction
    decoder := bin.NewBorshDecoder(txBytes)
    if err := tx.UnmarshalWithDecoder(decoder); err != nil {
        return "", fmt.Errorf("failed to unmarshal transaction: %w", err)
    }

    // 2. Fee Payerの秘密鍵を取得
    feePayerPrivateKey, err := solana.PrivateKeyFromBase58(feePayerPrivateKeyBase58)
    if err != nil {
        return "", fmt.Errorf("invalid fee payer private key: %w", err)
    }
    feePayerPubkey := feePayerPrivateKey.PublicKey()

    // 3. トランザクションのバリデーション
    // - Fee Payerがシステムウォレットであることを確認
    // Messageの最初のAccountKeyがFeePayer
    if len(tx.Message.AccountKeys) == 0 {
        return "", errors.New("transaction has no account keys")
    }
    if !tx.Message.AccountKeys[0].Equals(feePayerPubkey) {
        return "", errors.New("fee payer mismatch: transaction fee payer must be system wallet")
    }

    // - システム利用料（1 JPYC）が含まれているか確認

    foundSystemFee := false
    for _, inst := range tx.Message.Instructions {
        // SPL Token ProgramのInstructionを確認
        if int(inst.ProgramIDIndex) >= len(tx.Message.AccountKeys) {
            continue
        }
        programID := tx.Message.AccountKeys[inst.ProgramIDIndex]

        // Token Program IDを確認
        if programID.Equals(token.ProgramID) {
            // Transfer instructionを解析
            if len(inst.Data) < 1 {
                continue
            }
            instructionType := inst.Data[0]

            // Transfer instruction (3) を確認
            if instructionType == 3 {
                // Transfer instructionの構造:
                // - instructionType: 1 byte (3)
                // - amount: 8 bytes (uint64)
                if len(inst.Data) < 9 {
                    continue
                }
                amount := binary.LittleEndian.Uint64(inst.Data[1:9])

                // アカウントを確認（Transfer instructionは通常、source, destination, authorityの順）
                if len(inst.Accounts) >= 2 {
                    // destination accountを取得
                    if int(inst.Accounts[1]) < len(tx.Message.AccountKeys) {
                        destination := tx.Message.AccountKeys[inst.Accounts[1]]

                        // システム利用料ウォレットへの送金を確認
                        if destination.Equals(systemFeeWallet) && amount == expectedSystemFeeAmount {
                            foundSystemFee = true
                        }
                    }
                }
            }
        }
    }

    if !foundSystemFee {
        return "", fmt.Errorf("system fee not found: transaction must include %d lamports to system fee wallet", expectedSystemFeeAmount)
    }

    // 4. Fee Payerとして署名を追加
    // トランザクションに署名を追加（既存の署名は保持される）
    messageBytes, err := tx.Message.MarshalBinary()
    if err != nil {
        return "", fmt.Errorf("failed to marshal message: %w", err)
    }

    // Fee Payerの署名を追加（Signメソッドは2つの戻り値を返す）
    feePayerSignature, err := feePayerPrivateKey.Sign(messageBytes)
    if err != nil {
        return "", fmt.Errorf("failed to sign transaction: %w", err)
    }
    
    // Fee Payerの署名を最初の位置に設定（FeePayerは常に最初の署名）
    if len(tx.Signatures) == 0 {
        tx.Signatures = []solana.Signature{feePayerSignature}
    } else {
        tx.Signatures[0] = feePayerSignature
    }

    // 5. トランザクションを送信
    client := rpc.New(rpc.MainNetBeta_RPC)

    txBytes, err = tx.MarshalBinary()
    if err != nil {
        return "", fmt.Errorf("failed to marshal transaction: %w", err)
    }

    sig, err := client.SendTransaction(ctx, &tx)
    if err != nil {
        return "", fmt.Errorf("failed to send transaction: %w", err)
    }

    return sig.String(), nil
}

// ValidateJpycTransferInstruction JPYC送金Instructionを検証
func ValidateJpycTransferInstruction(
    inst solana.CompiledInstruction,
    accountKeys []solana.PublicKey,
    expectedTo solana.PublicKey,
    expectedAmount uint64,
) bool {
    if int(inst.ProgramIDIndex) >= len(accountKeys) {
        return false
    }
    programID := accountKeys[inst.ProgramIDIndex]

    // Token Program IDを確認
    if !programID.Equals(token.ProgramID) {
        return false
    }

    // Transfer instruction (3) を確認
    if len(inst.Data) < 9 || inst.Data[0] != 3 {
        return false
    }

    amount := binary.LittleEndian.Uint64(inst.Data[1:9])
    if amount != expectedAmount {
        return false
    }

    // destination accountを確認
    if len(inst.Accounts) < 2 {
        return false
    }
    if int(inst.Accounts[1]) >= len(accountKeys) {
        return false
    }
    destination := accountKeys[inst.Accounts[1]]

    return destination.Equals(expectedTo)
}
