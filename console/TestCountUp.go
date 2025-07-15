package console

import (
  "fmt"
  "log"

  "chat/common" // ← 実際のモジュール名に置き換えてね
)

func TestCountUp() {
  id := "1"

  // 1回だけカウントアップ
  next, err := common.CountUpID(id)
  if err != nil {
    log.Printf("CountUpID error: %v", err)
    return
  }

  fmt.Printf("New Count for ID '%s': %s\n", id, next)
}
