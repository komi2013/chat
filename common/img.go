package common

import (
  "encoding/base64"
  "io/ioutil"
  "log"
  "os"
  "strconv"
  "strings"
)

func AliasImgSave(aliasImg string, channelID string, i int ) string {
  imgPath := aliasImg
  if (strings.HasPrefix(aliasImg, "data:image")) {
    base64Data := strings.Split(aliasImg, ",")[1]
    imageData, err := base64.StdEncoding.DecodeString(base64Data)
    if err != nil {
        log.Println(err)
    }
    randPath := StringRand(4)
    dirPath := "/img/group/" + channelID + "/"
    os.MkdirAll("." + dirPath, 0755)
    imgPath = dirPath + strconv.Itoa(i) + "_" + randPath + ".png"
    filePath := "." + imgPath
    err = ioutil.WriteFile(filePath, imageData, 0644)
    if err != nil {
        log.Println(err)
    }
    log.Println("PNG image file saved successfully.")
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Println("Error getting file size:", err)
			return imgPath
		}
		fileSize := fileInfo.Size() // ファイルサイズを取得
		log.Println("fileSize", fileSize)
    // imgPath = ""
  }
  return imgPath
}
