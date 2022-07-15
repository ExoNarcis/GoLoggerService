package GoLogger

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ExoNarcis/GoNetReader"
)

type _PackFunction func(string) ([]byte, error)

type LoggerServiceClient struct {
	LocalLogPath       string
	NetLogNeed         bool
	NetServerIP        string
	NetServerPort      string
	LocalChannelSize   int
	NetChannelSize     int
	CritMassage        string
	PackFunction       _PackFunction
	_wLogchannel       chan string
	_generalLogChannel chan string
	_loggersService    sync.WaitGroup
	_netLogService     sync.WaitGroup
	_netError          bool
}

func (log *LoggerServiceClient) Disconnect() {
	if !log._netError {
		log._netError = true
		log.WriteLogs("Disconnect on demand")
		close(log._generalLogChannel)
	}
}

func (log *LoggerServiceClient) CloseLocalLog() {
	log.WriteLogs("Closed on demand")
	close(log._wLogchannel)
}
func (log *LoggerServiceClient) Exit() {
	log.Disconnect()
	log.CloseLocalLog()
}

func (log *LoggerServiceClient) logConnectionServer() error { // connect to log server
	conn, err := net.Dial("tcp", log.NetServerIP+":"+log.NetServerPort)
	if err != nil {
		go log.PrintAndWriteLocal("Connection to LOG server: " + log.NetServerIP + ":" + log.NetServerPort + " ERROR:" + err.Error() + "\t Only local log")
		log._netLogService.Done()
		log._loggersService.Done()
		log._netError = true
		return err
	}
	go log.PrintAndWriteLocal("connection to LOG server: " + log.NetServerIP + ":" + log.NetServerPort + "\t DONE!")
	if log.NetChannelSize == 0 {
		log.NetChannelSize = 5
	}
	log._generalLogChannel = make(chan string, log.NetChannelSize)
	go log.geneRalLogConnector(conn)
	return nil
}

func (log *LoggerServiceClient) geneRalLogConnector(Connection net.Conn) { // Sender
	defer Connection.Close()
	Connection.SetDeadline(time.Time{})
	log._netLogService.Done()
	for {
		Message, notclosed := <-log._generalLogChannel
		if notclosed {
			pack, err := log.PackFunction(Message)
			if err != nil {
				go log.WriteLocalLogs("Package convert error " + err.Error())
				continue
			}
			_, err = Connection.Write(GoNetReader.GetPackage(pack))
			if err != nil {
				close(log._generalLogChannel)
				log.WriteLocalLogs("Log Server Disconnect " + err.Error())
				log._netError = true
				//log.WriteLocalLogs("Retry 10 sec")
				//time.Sleep(10 * time.Second)
				//log._generalLogChannel <- Message + " netlog retry " + "[" + time.Now().Local().String() + "]"
				//log.logConnectionServer()
				return
			}
			//time.Sleep(time.Millisecond)
		} else {
			log._loggersService.Done()
			log._netError = true
			return
		}
	}
}

func (log *LoggerServiceClient) writer() { // local writer
	fileS, err := os.OpenFile(log.LocalLogPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer fileS.Close()
	if err != nil {
		log._loggersService.Done()
		return
	}
	for {
		Message, notclosed := <-log._wLogchannel
		if notclosed {
			if _, err := fileS.WriteString("[" + time.Now().Local().String() + "] " + Message + "\n"); err != nil {
				continue //?
			}
			if strings.Contains(Message, log.CritMassage) {
				close(log._wLogchannel)
				close(log._generalLogChannel)
				log._loggersService.Done()
				return
			}
		} else {
			log._loggersService.Done()
			_, notclosed = <-log._generalLogChannel
			if !notclosed {
				close(log._generalLogChannel)
			}
			return
		}
	}
}
func (log *LoggerServiceClient) WriteLocalLogs(Message string) { // write local
	log._wLogchannel <- Message
}
func (log *LoggerServiceClient) WriteLogs(Message string) { // write all if Net no Error state
	log._netLogService.Wait()
	log._wLogchannel <- Message
	if !log._netError {
		log._generalLogChannel <- Message
	}
}

func (log *LoggerServiceClient) PrintAndWrite(Message string) { // Prit and write logs
	fmt.Println(Message)
	log.WriteLogs(Message)
}

func (log *LoggerServiceClient) PrintAndWriteLocal(Message string) { // print and write local
	fmt.Println(Message)
	log.WriteLocalLogs(Message)
}

func (log *LoggerServiceClient) Wait() { // Wait func
	log._loggersService.Wait()
}
func (log *LoggerServiceClient) Init() { // init func
	log._loggersService.Add(2)
	log._netLogService.Add(1)
	if log.LocalChannelSize == 0 {
		log.LocalChannelSize = 5
	}
	log._wLogchannel = make(chan string, log.LocalChannelSize)
	if log.LocalLogPath != "" {
		go log.writer()
	}
	if log.NetLogNeed {
		if log.NetServerIP != "" && log.NetServerPort != "" {
			go log.logConnectionServer()
		}
	}
}
