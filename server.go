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

const messageSize = 20
const timeout = 5       //5 seconds
const timeToRun = 60    // 60 seconds

var protocolID = hash("Granada1.0")
var recievePort string
var sendPort string

//----------------------------------------------------------------------------
// TYPES
//----------------------------------------------------------------------------

// total 20 bytes
type message struct {
    ProtocolID  uint32 // 32 bits = 4 bytes
    Payload     string // 16 bytes 
}

type session struct {
    Conn     *net.UDPConn
    Address  *net.UDPAddr
}

type monitor struct {
    Conn *net.UDPConn
    Kill chan bool
}

//----------------------------------------------------------------------------
// MESSAGE METHODS
//----------------------------------------------------------------------------

func newMessage(payload string) *message {
    return &message{
        ProtocolID: protocolID,
        Payload: payload,
    }
}

//----------------------------------------------------------------------------
// SESSION METHODS
//----------------------------------------------------------------------------

func newSession(sendAddress string) *session {
    return &session{
        Conn: bindAddress(sendAddress, true),
    }
}

func (s *session) SendData(msg message) {
    // Encoding packet
    var messageBuffer bytes.Buffer
    err := gob.NewEncoder(&messageBuffer).Encode(msg); 
    checkErr(err)

    // Writing packet
    _, writeErr := s.Conn.Write(messageBuffer.Bytes())
    checkErr(writeErr)
}

//----------------------------------------------------------------------------
// MONITOR METHODS
//----------------------------------------------------------------------------

func newMonitor(address string) *monitor {
    return &monitor{
        Conn: bindAddress(address, false),
        Kill: make(chan bool),
    }
}

func (m *monitor) detectTimeOut(frame chan message, delay time.Duration) {
    buffer := make([]byte, messageSize)
    m.Conn.SetReadDeadline(time.Now().Add(delay))

    for {
        n, err := m.Conn.Read(buffer)
        checkErr(err)

        if n > 0 {
            // something was read before the deadline so reset the deadline
            var msg message
            err = gob.NewDecoder(bytes.NewReader(buffer[:n])).Decode(&msg)
            checkErr(err)

            if (msg.ProtocolID != protocolID) {
                fmt.Printf("not our protocol!\n");
            } else {
                frame <- msg
            }

            n = 0
            //go monitor.SendAck()
            m.Conn.SetReadDeadline(time.Now().Add(delay))
        }
        
        if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
            // Timeout error
            fmt.Println("No response")
            m.Kill <- true
            return
        }
    }
}
/*
func (m *monitor) SendAck() {
        msg := new(message)
        msg.ProtocolID = protocolID
        msg.Payload = "ack!"

        // Encoding packet
        var messageBuffer bytes.Buffer
        err := gob.NewEncoder(&messageBuffer).Encode(msg); 
        checkErr(err)

        // Writing packet
        _, writeErr := m.Conn.WriteToUDP(messageBuffer.Bytes())
        checkErr(writeErr)
}
*/
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


 /*
 potential decode function?
                //decoding response
                var value message
                err = gob.NewDecoder(bytes.NewReader(buf[:n])).Decode(&value)
                checkErr(err)

                if (value.ProtocolID != protocolID) {
                    fmt.Printf("not our protocol!\n");
                } else {
                    fmt.Printf("recieved %s from %v\n", value.Payload, remoteaddr)
                }
                */


func SendThread(frame chan *message) {
    var addressBuffer bytes.Buffer
    addressBuffer.WriteString(os.Args[1])
    addressBuffer.WriteString(sendPort)
    session := newSession(addressBuffer.String())

    for {
        select {
            case msgPtr := <-frame:
                msg := *msgPtr
                fmt.Printf("listen thread recieved %+v through the outChannel\n", msg)
                session.SendData(msg)
        }
    }
}

func ListenThread(frame chan message) {
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

/*
func KillConnection(<- done chan bool)  {
    timer := time.NewTimer(time.Second * timeToRun)
    <- timer.C
    done <- true
}

func ControlCKill() {
    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt)

    select {
    case sig := <-c:
        fmt.Printf("Got %s signal. Aborting...\n", sig)
        os.Exit(1)
    }
}
*/
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

    //doneChannel := make(chan bool, 2)

    // runs the program for a certain amount of time
    //go KillGame(doneChannel)
    //go ControlCKill()

    frameChannelOut := make(chan *message, 5)
    //  frameChannelIn := make(chan message, 5)

    go func() {
        for {
            msg := newMessage("whats up!")
            frameChannelOut <- msg
            fmt.Printf("sent message\n");
            time.Sleep(time.Second)
        }
    }()

    go SendThread(frameChannelOut)
    /*
    go ListenThread(frameChannelIn)

    fmt.Printf("Waiting for a response, and sending data\n")

    for {
        msg := <-frameChannelIn
        fmt.Printf("recieved message: %+v\n", msg)
    }
    */
    for {
        time.Sleep(time.Second);
    }
}


/*
application needs to:

listen at  a port, read incoming data in loop and pass it somewhere(buffer?) when it gets here
send data every second (start stupid then make a frame/packet channel to put data into)
*/
