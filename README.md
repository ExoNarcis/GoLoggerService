# GoLoggerService
## Go Local Logger and Net over TCP
### Example:
#### Client:
```
package main
import (
  "github.com/ExoNarcis/GoLogger"
)

type NetPackage struct {
	Name   string
	Body   NetPackageBody
	Sender string
	Time   int64
}

type NetPackageBody struct {
	Command  string
	PackBody string
}

var Logger = GoLogger.LoggerServiceClient{LocalLogPath: "log.txt",
	NetLogNeed:       true, // if you need net log
	NetServerIP:      "127.0.0.1", // net ip
	NetServerPort:    "1338", // target port 
	LocalChannelSize: 10, 
	NetChannelSize:   10,
	CritMassage:      "CRIT!"}
  
func PackFun(pack string) ([]byte, error) {
  unJsonPack := NetPackage{Name: "NetLog",
		Body:   NetPackageBody{Command: "GoTCPMain NetLog Write", PackBody: "GoTCPMain:" + pack},
		Sender: "GoTCPMain",
		Time:   time.Now().Unix()}

  jpack, errJ := json.Marshal(unJsonPack)
	if errJ != nil {
		return []byte(""), errJ
	}
	return jpack, nil
}
  
func main() {
  Logger.PackFunction = PackFun
  Logger.Init()
  defer Logger.Wait()
  go Logger.WriteLogs(" NET + LOCAL")
  go Logger.WriteLocalLogs("ONLY LOCAL")
}
```

#### Server:
```
package main

import (
	"GoLogger"
	"encoding/json"
)

type NetPackage struct {
	Name   string
	Body   NetPackageBody
	Sender string
	Time   int64
}

type NetPackageBody struct {
	Command  string
	PackBody string
}

var NetLogServer = GoLogger.LoggerServiceServer{LocalLogPath: "log.txt", // local log path
	NetLogPath:        "GeneRalNetLog.txt", // Net Log Path
	NetServerProtocol: "tcp", // type (Net.Conn)
	NetServerPort:     ":1338", // port 
	LocalChannelSize:  10, 
	NetChannelSize:    10,
	CritMassage:       "CRIT!"}

func main() {
	NetLogServer.UnPackFunction = UnPackFunc
	NetLogServer.Init()
	defer NetLogServer.Wait()
}

func UnPackFunc(mess string) (string, error) {
	UnJPack := NetPackage{}
	errUnJ := json.Unmarshal([]byte(mess), &UnJPack)
	if errUnJ != nil {
		return "", errUnJ
	}
	return UnJPack.Body.PackBody, nil
}

```
