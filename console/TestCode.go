package console

import (
  "fmt"
  // "log"

  "chat/common" // ← 実際のモジュール名に置き換えてね
)

func TestCode() {
  // id, err := common.CountUpID("wow")
  // if err != nil {
  //   fmt.Printf("CountUpID error: %v", err)
  //   return
  // }
  // fmt.Printf("New Count for ID '%s': %s\n", id)

	// addr, err := common.GenerateInvoiceAddress("12345") // InvoiceID
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println("Invoice Address =", addr)

  cfg := common.LoadConfig()


	fmt.Println("GoPort =", cfg.GoPort)


}
