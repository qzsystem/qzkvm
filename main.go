package main

import (
   "fmt"
   "net/http"
   "qzkvm/src/pkg/logging"
   "qzkvm/src/pkg/setting"
   "qzkvm/src/qzcloud"
   "qzkvm/src/routers"
)

func init()  {
   qzcloud.Setup()
}
func main() {
   router := routers.InitRouter()

   s := &http.Server{
      Addr:           fmt.Sprintf(":%d", setting.HTTPPort),
      Handler:        router,
      ReadTimeout:    setting.ReadTimeout,
      WriteTimeout:   setting.WriteTimeout,
      MaxHeaderBytes: 1 << 20,
   }
   logging.Info("[info] start http server listening %s", setting.HTTPPort)
   s.ListenAndServe()
}