package main

import (
        "fmt"
        "net"
        "hash/fnv"
        "encoding/gob"
        "bytes"
        "os"
        "time"
        "strconv"
)

//----------------------------------------------------------------------------
// GLOBALS
//----------------------------------------------------------------------------

const MessageSize = 48
const timeout = 5       //5 seconds
const timeToRun = 60    // 60 seconds

var protocolID = hash("Granada1.0")
var recievePort string
var sendPort string

//----------------------------------------------------------------------------
// TYPES
//----------------------------------------------------------------------------

// total 48 bytes
type Message struct {
    ProtocolID  uint32 // 32 bits = 4 bytes
    Payload     [44]byte // 16 bytes 
}

type session struct {
    Conn     *net.UDPConn
    Address  *net.UDPAddr
    Encoder  *gob.Encoder
}

type monitor struct {
    Conn *net.UDPConn
    Kill chan bool
    Decoder *gob.Decoder
}

//----------------------------------------------------------------------------
// MESSAGE METHODS
//----------------------------------------------------------------------------

func newMessage(payload string) *Message {
    var payloadBuffer [44]byte
    copy(payloadBuffer[:], payload)

    return &Message{
        ProtocolID: protocolID,
        Payload: payloadBuffer,
    }
}

//----------------------------------------------------------------------------
// SESSION METHODS
//----------------------------------------------------------------------------

func newSession(sendAddress string) *session {
    conn := bindAddress(sendAddress, true)

    return &session{
        Conn: conn,
        Encoder: gob.NewEncoder(conn),
    }
}

//----------------------------------------------------------------------------
// MONITOR METHODS
//----------------------------------------------------------------------------

func newMonitor(address string) *monitor {
    conn := bindAddress(address, false)

    return &monitor{
        Conn: conn,
        Kill: make(chan bool),
        Decoder: gob.NewDecoder(conn),
    }
}

func (m *monitor) detectTimeOut(frame chan Message, delay time.Duration) {   
    m.Conn.SetReadDeadline(time.Now().Add(delay))

    for {
        var msg Message
        err := m.Decoder.Decode(&msg)

        if (msg.ProtocolID != protocolID) {
            fmt.Printf("not our protocol!\n");
        } else {
            frame <- msg
        }

        m.Conn.SetReadDeadline(time.Now().Add(delay))
        
        if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
            // Timeout error
            fmt.Println("No response")
            m.Kill <- true
            return
        }
    }
}

//----------------------------------------------------------------------------
// FUNCTIONS
//----------------------------------------------------------------------------

func hash(s string) uint32 {
        h := fnv.New32a()
        h.Write([]byte(s))
        return h.Sum32()
}

func checkErr(err error) {
    if err != nil {
        fmt.Println("ERROR: ", err)
        os.Exit(1)
    }
}

func bindAddress(address string, dial bool) *net.UDPConn {
    var conn *net.UDPConn

    if (dial) {
        // dial
        var dialErr error
        addr, err := net.ResolveUDPAddr("udp", address)
        checkErr(err)

        conn, dialErr = net.DialUDP("udp", nil, addr)
        checkErr(dialErr)
    } else {
        // listen
        var listenErr error
        addr, err := net.ResolveUDPAddr("udp", address)
        checkErr(err)

        conn, listenErr = net.ListenUDP("udp", addr)
        checkErr(listenErr)
    }

    return conn
}


func SendThread(frame chan *Message) {
    var addressBuffer bytes.Buffer
    addressBuffer.WriteString(os.Args[1])
    addressBuffer.WriteString(sendPort)
    session := newSession(addressBuffer.String())

    for {
        select {
            case msgPtr := <-frame:
                msg := *msgPtr
                session.Encoder.Encode(msg)
        }
    }
}

func ListenThread(frame chan Message) {
    monitor := newMonitor(recievePort)
    defer monitor.Conn.Close()
    monitor.detectTimeOut(frame, timeout)

    for {
        select {
        case <- monitor.Kill:
            //game is killed
            fmt.Println("Stopped listening")
            return
        }
    }
}

//----------------------------------------------------------------------------
// MAIN
//----------------------------------------------------------------------------

func main() {
    portBool, parseErr := strconv.ParseBool(os.Args[2])
    checkErr(parseErr)

    if (portBool) {
        sendPort = ":10001"
        recievePort = ":10002"
    } else {
        sendPort = ":10002"
        recievePort = ":10001"
    }

    frameChannelOut := make(chan *Message, 5)
    frameChannelIn := make(chan Message, 5)

    go func() {
        for {
            msg := newMessage("whats up!")
            frameChannelOut <- msg
            fmt.Printf("sent Message\n");
            time.Sleep(time.Second)
        }
    }()

    go SendThread(frameChannelOut)
    go ListenThread(frameChannelIn)

    fmt.Printf("Waiting for a response, and sending data\n")

    for {
        msg := <-frameChannelIn
        fmt.Printf("recieved Message: %+v\n", msg)
    }
}
