package main

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

func main() {
  const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
  r := gin.Default()

  canvasCtx := make(map[string][]byte)

  //melody
  m := melody.New()

  //routing
  r.GET("/", func(c *gin.Context) {
    newFlag := c.Query("new")
    if(newFlag == "true") {
      rand.Seed(time.Now().UnixNano())
      digit := rand.Intn(4)+3
      b := make([]byte, int(digit))
      rand.Read(b)

      var result string
      f := true
      for f {
        for _, v := range b {
          result += string(letters[int(v)%len(letters)])
        }
        if _, ok := canvasCtx[result]; !ok {
          f = false
        }
      }
      canvasCtx[result] = []byte{0}
      c.Redirect(http.StatusSeeOther, "/?session=" + result)
      return
    }

    session := c.Query("session")
    if _, ok := canvasCtx[session]; ok && session != "" {
      http.ServeFile(c.Writer, c.Request, "index.html")
    } else {
      http.ServeFile(c.Writer, c.Request, "enter.html")
    }
  })

  r.POST("/store", func(c *gin.Context) {
    img, _ := c.FormFile("image_file")
    png, _ := img.Open()
    defer png.Close()

    canvasCtx[c.Query("session")], _ = ioutil.ReadAll(png)
  })

  r.GET("/restore", func(c *gin.Context) {
    c.Data(http.StatusOK, "image/png", canvasCtx[c.Query("session")])
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
        return s1 == s2 && s != ss
      })

    case "store":
      m.BroadcastFilter([]byte("restore:"), func (ss *melody.Session) bool {
        s1, _ := s.Get("session")
        s2, _ := ss.Get("session")
        return s1 == s2 && s != ss
      })

    }
  })

  r.Run(":8888")
}