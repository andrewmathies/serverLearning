package main

import (
    "fmt"
    "net"
	"hash/fnv"
	"encoding/gob"
	"bytes"
	"os"
)

// total 20 bytes
type message struct {
	ProtocolID uint32 // 32 bits = 4 bytes
	Payload [16]byte // 16 bytes 
}

// 20 bytes * 8 bits = 128
var messageSize = 160
var protocolID = hash("Granada1.0")

func hash(s string) uint32 {
        h := fnv.New32a()
        h.Write([]byte(s))
        return h.Sum32()
}

func checkErr(err error) {
    if err != nil {
        fmt.Println("ERROR:", err)
        os.Exit(1)
    }
}

func main() {
	// Initialization

	/*
	localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	checkErr(err)
	*/

	serverAddr, addrErr := net.ResolveUDPAddr("udp", "13.58.214.125:10001")
	checkErr(addrErr)

    conn, dialErr := net.DialUDP("udp", nil, serverAddr)
    checkErr(dialErr)

	defer conn.Close()
	fmt.Printf("initialized\n")

	// Building packet
	var buffer [16]byte
	copy(buffer[:], "hello server!")

	msg := new(message)
	msg.ProtocolID = protocolID
	msg.Payload = buffer

	// Encoding packet
	var convertedMsg bytes.Buffer
	err := gob.NewEncoder(&convertedMsg).Encode(msg); 
	checkErr(err)

	// Writing packet
	_, writeErr := conn.Write(convertedMsg.Bytes())
	checkErr(writeErr)
	
	// Reading response
	readBuf := make([]byte, messageSize)
	n, err := conn.Read(readBuf)
	checkErr(err)
	var value message
	err = gob.NewDecoder(bytes.NewReader(readBuf[:n])).Decode(&value)
	checkErr(err)

	if (value.ProtocolID != protocolID) {
		fmt.Printf("not our protocol!\n");
	} else {
		fmt.Printf("recieved %s from server\n", value.Payload)
	}
	
}
