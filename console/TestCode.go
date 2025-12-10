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

  cfg := common.LoadConsoleConfig()

	fmt.Println("GoPort =", cfg.GoPort)

	appID := "c0b2a419-9c3c-408f-ab32-671407d6e3ad"
	secret := "E3bVL++TkEprPjK5+DNYdwxuw7yawPv8o1OsBl4pPTU="

	// Token 有効期限（日単位）
	token, err := common.GenerateSkyWayToken(appID, secret)
	if err != nil {
		fmt.Println("Skyway token error:", err)
		return
	}

	fmt.Println("SkyWay Token:")
	fmt.Println(token)

}
