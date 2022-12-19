package main

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

func main() {
  r := gin.Default()

  canvasCtx := make(map[string][]byte)

  //melody
  m := melody.New()

  //routing
  r.GET("/", func(c *gin.Context) {
    http.ServeFile(c.Writer, c.Request, "index.html")
  })

  r.POST("/store", func(c *gin.Context) {
    img, _ := c.FormFile("image_file")
    png, _ := img.Open()
    defer png.Close()

    canvasCtx["1"], _ = ioutil.ReadAll(png)
  })

  r.GET("/restore", func(c *gin.Context) {
    c.Data(http.StatusOK, "image/png", canvasCtx["1"])
  })

  r.GET("/ws", func(c *gin.Context) {
    m.HandleRequest(c.Writer, c.Request)
  })

  m.HandleMessage(func(s *melody.Session, msg []byte) {
    msgArray := strings.SplitN(string(msg), ":", 2)

    switch msgArray[0] {
    case "change":
      s.Set("session", msgArray[1])
    
    case "draw":
      m.BroadcastFilter(msg, func (ss *melody.Session) bool {
        s1, _ := s.Get("session")
        s2, _ := ss.Get("session")
        return s1 == s2
      })

    }
  })

  r.Run(":8000")
}