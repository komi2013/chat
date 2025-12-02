// copy-after-build.js  (CommonJS版)
console.log("✅ start");
const fs = require('fs');
const path = require('path');

// dist の場所
const distAssets = path.join(__dirname, './dist/assets');
const distView = path.join(__dirname, './dist/view');

// コピー先
const publicAssets = path.join(__dirname, '../../assets');
const viewDir = path.join(__dirname, '../../view');

// ディレクトリ削除関数
function removeDir(dirPath) {
  if (fs.existsSync(dirPath)) {
    fs.rmSync(dirPath, { recursive: true, force: true });
  }
}

// ディレクトリ作成関数
function ensureDir(dirPath) {
  if (!fs.existsSync(dirPath)) {
    fs.mkdirSync(dirPath, { recursive: true });
  }
}

// ディレクトリコピー関数
function copyDir(src, dest) {
  if (fs.existsSync(src)) {
    fs.cpSync(src, dest, { recursive: true });
  }
}

// ---- 実行 ----

// public/assets をクリーン
removeDir(publicAssets);
ensureDir(publicAssets);
copyDir(distAssets, publicAssets);

// view をクリーン
removeDir(viewDir);
ensureDir(viewDir);
copyDir(distView, viewDir);

console.log("✅ Assets & View copied successfully!");
