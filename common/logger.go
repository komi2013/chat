package common

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type DailyLogger struct {
	mu       sync.Mutex
	prefix   string
	dir      string
	currDate string
	logger   *log.Logger
	file     *os.File
}

// コンストラクタ
func NewDailyLogger(prefix string) *DailyLogger {
	return &DailyLogger{
		prefix: prefix,
		dir:    "./log",
	}
}

// ロガーを取得（必要ならファイルをローテーション）
func (dl *DailyLogger) getLogger() *log.Logger {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	if dl.logger == nil || dl.currDate != today {
		// 古いファイルを閉じる
		if dl.file != nil {
			_ = dl.file.Close()
		}

		_ = os.MkdirAll(dl.dir, 0755)
		logFileName := dl.prefix + today + ".log"
		logPath := filepath.Join(dl.dir, logFileName)

		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("ログファイルを開けません: %v", err)
		}

		dl.file = file
		dl.logger = log.New(file, "[CLIENT-ERROR] ", log.LstdFlags|log.Lshortfile)
		dl.currDate = today
	}

	return dl.logger
}

// ログ出力（Printfスタイル）
func (dl *DailyLogger) Printf(format string, v ...interface{}) {
	dl.getLogger().Printf(format, v...)
}

func LogError(message string, err error, errs ...interface{}) {
	funcName := getControllerFuncName()
	log.Printf("[LogError] %s: %s - %v %v", funcName, message, err, errs)
}

func getControllerFuncName() string {
	var lastValidFunc string // 直前の関数を保存する
	for i := 2; i < 15; i++ { // 2 〜 15 の範囲で関数名を探る
		pc, _, _, ok := runtime.Caller(i)
		if !ok {
			break // これ以上スタックがない場合はループ終了
		}
		funcName := runtime.FuncForPC(pc).Name()

		// net/http の関数なら、直前の関数を返す
		if strings.Contains(funcName, "net/http.") {
			return lastValidFunc
		}

		// runtime や net/http でない関数を lastValidFunc に保存
		if !strings.Contains(funcName, "runtime.") {
			lastValidFunc = funcName
		}
	}
	return "unknown"
}
