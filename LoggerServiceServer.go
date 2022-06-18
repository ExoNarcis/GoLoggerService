package GoLogger

import (
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ExoNarcis/GoNetReader"
)

type _UnPackFunction func(string) (string, error)

type LoggerServiceServer struct {
	LocalLogPath       string
	NetLogPath         string
	NetServerProtocol  string
	NetServerPort      string
	LocalChannelSize   int
	NetChannelSize     int
	CritMassage        string
	UnPackFunction     _UnPackFunction
	_wLogchannel       chan string
	_generalLogChannel chan string
	_critChannelLog    sync.WaitGroup
}

func (log *LoggerServiceServer) netGeneRalWriter() { // NetLog Writer
	fileS, err := os.OpenFile(log.NetLogPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer fileS.Close()
	if err != nil {
		return
	}
	for {
		Message, notclosed := <-log._generalLogChannel
		if notclosed {
			if _, err := fileS.WriteString("[" + time.Now().Local().String() + "] " + Message + "\n"); err != nil {
				continue //?
			}
		} else {
			return
		}
	}
}

func (log *LoggerServiceServer) NetLogWrite(mess string) { // Write to channel
	log._generalLogChannel <- mess
}

func (log *LoggerServiceServer) writer() { // writer local
	fileS, err := os.OpenFile(log.LocalLogPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer fileS.Close()
	if err != nil {
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
				log._critChannelLog.Done()
				return
			}
		} else {
			log._critChannelLog.Done()
			return
		}
	}
}

func (log *LoggerServiceServer) WriteLocalLogs(Message string) { // local Writer
	log._wLogchannel <- Message
}

func (log *LoggerServiceServer) Init() { // init func
	log._critChannelLog.Add(1)
	log._wLogchannel = make(chan string, log.LocalChannelSize)
	log._generalLogChannel = make(chan string, log.NetChannelSize)
	go log.writer()
	go log.netGeneRalWriter()
	go log.serverMain()
}

func (log *LoggerServiceServer) serverMain() { // Main Serv
	ListenS, err := net.Listen(log.NetServerProtocol, log.NetServerPort)
	if err != nil {
		go log.WriteLocalLogs("CRIT! Error start Listener, Configuration: protocol:" + log.NetServerProtocol + " localport:" + log.NetServerPort + " Error:" + err.Error())
		return
	}
	go log.WriteLocalLogs("\t\t\t STARTING SERVER \t\t\t\n Starting Listener: Protocol:" + log.NetServerProtocol + " Port:" + log.NetServerPort)
	defer ListenS.Close()
	for {
		connection, err := ListenS.Accept()
		if err != nil {
			go log.WriteLocalLogs("Error Listener potocol:" + log.NetServerProtocol + " localport: " + log.NetServerPort)
		}
		go log.WriteLocalLogs("New Connection : " + connection.RemoteAddr().String())
		go log.connectionRouter(connection)
	}
}
func (log *LoggerServiceServer) connectionRouter(Connection net.Conn) { // Router
	defer Connection.Close()
	reader := GoNetReader.NewNetReader()
	for {
		Pack, err := reader.NetRead(Connection)
		Unpack := ""
		if err != nil {
			if err == io.EOF {
				go log.WriteLocalLogs("Connection " + Connection.RemoteAddr().String() + " Closed")
				return
			}
			go log.WriteLocalLogs("Package Error: " + Connection.RemoteAddr().String() + " ERROR: " + err.Error())
			continue
		}
		Unpack, err = log.UnPackFunction(Pack)
		if err != nil {
			go log.WriteLocalLogs("Error " + err.Error() + " to Deserialse Message: " + Pack + " Invalid")
		}
		go log.NetLogWrite(Unpack)
	}
}

func (log *LoggerServiceServer) Wait() { // Wait func
	log._critChannelLog.Wait()
}
