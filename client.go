package main

import (
    "fmt"
    "net"
    "time"
	"bufio"
	"hash/fnv"
	"encoding/gob"
	"bytes"
)

// total 20 bytes
type message struct {
	ProtocolID uint32 // 32 bits = 4 bytes
	Payload [16]byte // 16 bytes 
}

func hash(s string) uint32 {
        h := fnv.New32a()
        h.Write([]byte(s))
        return h.Sum32()
}

// 20 bytes * 8 bits = 128
var messageSize = 160

func main() {
	//INITIALIZATION

	/*
	localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")

	if err != nil {
		fmt.Printf("some error %v\n", err)
		return
	}
	*/

	serverAddr, addrErr := net.ResolveUDPAddr("udp", "13.58.214.125:10001")

	if addrErr != nil {
		fmt.Printf("some error %v\n", addrErr)
		return
	}

    conn, dialErr := net.DialUDP("udp", nil, serverAddr)

    if dialErr != nil {
            fmt.Printf("some error %v\n", dialErr)
            return
    }

	defer conn.Close()
	fmt.Printf("initialized\n")

	// Building packet

	msg := new(message)

	msg.ProtocolID = hash("Granada1.0")
	fmt.Printf("Our ProtocolID is: %v\n", msg.ProtocolID)
	var buffer [16]byte
	copy(buffer[:], "hello server!")
	msg.Payload = buffer

	// Encoding packet
	var convertedMsg bytes.Buffer
	if err := gob.NewEncoder(&convertedMsg).Encode(msg); err != nil {
		fmt.Printf("encode err: ", err)
		return
	}

	// Writing packet
	_, writeErr := conn.Write(convertedMsg.Bytes())
	
	if writeErr != nil {
		fmt.Printf("write error ", writeErr)
		return
	}
	
	// Reading response
	readBuf := make([]byte, messageSize)
	placeholder, readErr := bufio.NewReader(conn).Read(readBuf)

	if readErr != nil {
		fmt.Printf("error reading! ", readErr)
		fmt.Printf("plus this! ", placeholder)
		return
	} else {
		fmt.Printf("%s\n\n", readBuf)
	}

	time.Sleep(time.Second * 1)
}		

/*
        fmt.Fprintf(conn, "Hi UDP Server, How are you doing?")
        _, err = bufio.NewReader(conn).Read(p)
        if err == nil {
                fmt.Printf("%s\n", p)
        } else {
                fmt.Printf("Some error %v\n", err)
        }
        conn.Close()
}
*/
