package GoLogger

import (
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
	fileS, errO := os.OpenFile(log.NetLogPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer fileS.Close()
	if errO != nil {
		return
	}
	for {
		Message, err := <-log._generalLogChannel
		if err {
			if _, errW := fileS.WriteString("[" + time.Now().Local().String() + "] " + Message + "\n"); errW != nil {
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
	fileS, errO := os.OpenFile(log.LocalLogPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer fileS.Close()
	if errO != nil {
		return
	}
	for {
		Message, err := <-log._wLogchannel
		if err {
			if _, errW := fileS.WriteString("[" + time.Now().Local().String() + "] " + Message + "\n"); errW != nil {
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
	ListenS, ErrorList := net.Listen(log.NetServerProtocol, log.NetServerPort)
	if ErrorList != nil {
		go log.WriteLocalLogs("CRIT! Error start Listener, Configuration: protocol:" + log.NetServerProtocol + " localport:" + log.NetServerPort + " Error:" + ErrorList.Error())
		return
	}
	go log.WriteLocalLogs("\t\t\t STARTING SERVER \t\t\t\n Starting Listener: Protocol:" + log.NetServerProtocol + " Port:" + log.NetServerPort)
	defer ListenS.Close()
	for {
		connection, errConnect := ListenS.Accept()
		if errConnect != nil {
			go log.WriteLocalLogs("Error Listener potocol:" + log.NetServerProtocol + " localport: " + log.NetServerPort)
		}
		go log.WriteLocalLogs("New Connection : " + connection.RemoteAddr().String())
		go log.connectionRouter(connection)
	}
}
func (log *LoggerServiceServer) connectionRouter(Connection net.Conn) { // Router
	reader := GoNetReader.NewNetReader()
	for {
		Pack, ErrRead := reader.NetRead(Connection)
		if ErrRead != nil {
			go log.WriteLocalLogs("Package Error: " + Connection.RemoteAddr().String() + " ERROR: " + ErrRead.Error())
			Connection.Close()
			return
		}
		UnJPack, errUnJ := log.UnPackFunction(Pack)
		if errUnJ != nil {
			go log.WriteLocalLogs("Error " + errUnJ.Error() + " to Deserialse Message: " + Pack + " Invalid")
		}
		go log.NetLogWrite(UnJPack)
	}
}

func (log *LoggerServiceServer) Wait() { // Wait func
	log._critChannelLog.Wait()
}
