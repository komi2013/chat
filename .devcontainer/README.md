# Dev Container: Go + Vue + Android + Cline (isolated)

このプロジェクト (`chat`) を **VS Code / Cline / Go / Node.js / Android SDK (adb・Gradle)**
が揃った単一コンテナで開発するための構成です。
ホスト PC の秘密情報 (`~/.ssh` 等) は **一切マウントしません**。

```
chat/
 └── .devcontainer/
      ├── devcontainer.json   # コンテナ定義（拡張機能・adb 転送・隔離）
      ├── Dockerfile          # Go / Node / Java17 / Android SDK のツールチェーン
      └── README.md           # このファイル
```

## 収録ツールチェーン

| ツール | バージョン | 根拠 |
| --- | --- | --- |
| Go | 1.23.x | `go.mod` (`go 1.23.0`) |
| Node.js / npm | 20 LTS | `.github/workflows/deploy.yml` |
| Java | OpenJDK 17 | Android Gradle Plugin 8.2 の要件 |
| Android SDK | platform-tools(adb) / platforms;android-34 / build-tools;34.0.0 | `android/app/build.gradle` |
| Gradle | 8.5 (wrapper) | `android/gradle/wrapper/gradle-wrapper.properties` |

## 🚀 セットアップ手順

### 1. ホスト PC 側で ADB サーバーを起動

Android 実機を USB 接続し、実機側で「USB デバッグ」を許可します。
その後、**コンテナ外（ホスト PC）のターミナル** で実行します
（このターミナルは開いたままにしておきます）。

```bash
# 既存の adb サーバーを停止
adb kill-server

# 全インターフェースから接続を受け付ける形で起動（Docker からの接続を許可）
adb -a nodaemon server
```

> Windows の場合は、このターミナルを **管理者権限** で開くと確実です。

### 2. コンテナを起動

1. VS Code でプロジェクトフォルダを開く。
2. コマンドパレット（`Ctrl+Shift+P` / `Cmd+Shift+P`）を開く。
3. **`Dev Containers: Reopen in Container`** を選択。
4. 初回はイメージのビルドに数分かかります。完了するとコンテナ内の VS Code が開きます。

### 3. 動作確認（コンテナ内ターミナル / Cline）

```bash
# Go
go version            # go1.23.x
go build ./...        # バックエンドのビルド

# Node.js (Vue / Vite)
cd vue
npm ci
npm run build

# Android 実機 (adb -> ホストの ADB サーバー経由)
adb devices           # ホストに接続した実機のシリアルが表示される
adb logcat            # 実機のログが流れる（Ctrl+C で停止）
```

Android アプリのビルド（Gradle）:

```bash
cd android
./gradlew assembleDebug
# 実機インストール: ./gradlew installDebug  （adb はホストへ自動転送される）
```

## 🔒 セキュリティ / 隔離について

- `devcontainer.json` に `mounts` を **一切定義していない** ため、
  ホストの `~/.ssh`・`~/.gitconfig`・個人ファイルはコンテナ内に存在しません。
  コンテナ内で AI エージェントが全検索しても、ホストの鍵には到達できません。
- マウントされるのは VS Code が開く **ワークスペース（このプロジェクトフォルダ）だけ** です。
- Git 認証が必要な場合は、コンテナ内で **PAT (Personal Access Token)** を使うか、
  コンテナ専用の SSH 鍵をコンテナ内で生成して使用してください。
- `adb` は USB に直接アクセスせず、ホストの ADB サーバー（TCP 5037）へ
  `ADB_SERVER_SOCKET=tcp:host.docker.internal:5037` で転送するだけです。

## 🔧 トラブルシューティング

- **`adb devices` に実機が出ない**
  - ホスト側で `adb kill-server` → `adb -a nodaemon server` で起動し直しているか確認。
  - 実機に「USB デバッグを許可しますか？」のポップアップが出ていないか確認。
  - Windows で失敗する場合はホストのターミナルを管理者権限で実行。
- **`host.docker.internal` が解決できない**
  - `runArgs` の `--add-host=host.docker.internal:host-gateway` が有効か確認（Docker 20.10+）。
- **`java` / `sdkmanager` が見つからない**
  - イメージのビルドが失敗していないか確認：
    `docker build -t chat-dev .devcontainer`
- **sdkmanager のライセンス再取得**
  - コンテナ内で `sdkmanager --licenses` を再実行して `y` を入力。

## 💡 任意: 再ビルドを速くする（ホスト非依存の名前付きボリューム）

隔離を保ったまま（＝ホストの実ファイルは見せず）キャッシュを持たせる場合のみ、
`devcontainer.json` に名前付きボリュームを追加できます（ホストパスは露出しません）。

```jsonc
"mounts": [
  "source=chat-go-mod,target=/go/pkg/mod,type=volume",
  "source=chat-go-build,target=/home/vscode/.cache/go-build,type=volume",
  "source=chat-gradle,target=/home/vscode/.gradle,type=volume",
  "source=chat-node,target=/workspaces/chat/vue/node_modules,type=volume"
]
```
