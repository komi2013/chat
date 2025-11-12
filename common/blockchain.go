package common

import (
    "crypto/ecdsa"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "math/big"
    "net/http"
    "strings"

    "github.com/ethereum/go-ethereum/crypto"
    hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
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
