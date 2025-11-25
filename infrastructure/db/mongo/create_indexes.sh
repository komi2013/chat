#!/bin/bash
# MongoDB インデックス作成スクリプト (シェルスクリプト)
# 環境変数から接続文字列を取得してJavaScriptスクリプトを実行

# 環境変数の確認
if [ -z "$MONGO_SESSION" ] || [ -z "$MONGO_NICKNAME" ] || [ -z "$MONGO_AD" ] || [ -z "$MONGO_USER" ] || [ -z "$MONGO_INVOICE" ]; then
  echo "エラー: 必要な環境変数が設定されていません"
  echo "以下の環境変数を設定してください:"
  echo "  MONGO_SESSION"
  echo "  MONGO_NICKNAME"
  echo "  MONGO_AD"
  echo "  MONGO_USER"
  echo "  MONGO_INVOICE"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
JS_SCRIPT="$SCRIPT_DIR/create_indexes.js"

# MongoDBコマンドの確認
if command -v mongosh &> /dev/null; then
  MONGO_CMD="mongosh"
elif command -v mongo &> /dev/null; then
  MONGO_CMD="mongo"
else
  echo "エラー: mongosh または mongo コマンドが見つかりません"
  exit 1
fi

echo "=== MongoDB インデックス作成スクリプト ==="
echo "使用コマンド: $MONGO_CMD"
echo ""

# 各データベースに対してスクリプトを実行
# 注意: このスクリプトは各データベースに個別に接続する必要があります
# JavaScriptスクリプト内でgetSiblingDBを使用しているため、最初の接続文字列を使用

# 最初のデータベース（session）に接続してスクリプトを実行
echo "接続中: $MONGO_SESSION"
$MONGO_CMD "$MONGO_SESSION" "$JS_SCRIPT"

if [ $? -eq 0 ]; then
  echo ""
  echo "✓ インデックス作成が完了しました"
else
  echo ""
  echo "✗ エラーが発生しました"
  exit 1
fi

