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

func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
        // Building packet
        var buffer [16]byte
        copy(buffer[:], "hello client!")

        msg := new(message)
        msg.ProtocolID = hash("Granada1.0")
        msg.Payload = buffer

        // Encoding packet
        var convertedMsg bytes.Buffer
        err := gob.NewEncoder(&convertedMsg).Encode(msg); 
        checkErr(err)

        // Writing packet
        _, writeErr := conn.WriteToUDP(convertedMsg.Bytes(), addr)
        checkErr(writeErr)
}

func main() {
        buf := make([]byte, messageSize)

        serverAddr, err := net.ResolveUDPAddr("udp", ":10001")
        checkErr(err)

        serverConn, err := net.ListenUDP("udp", serverAddr)
        checkErr(err)
        
        defer serverConn.Close()
        fmt.Printf("Running, waiting for a response!\n")

        for {
                n, remoteaddr, err := serverConn.ReadFromUDP(buf)
                checkErr(err)

                var value message
                err = gob.NewDecoder(bytes.NewReader(buf[:n])).Decode(&value)
                checkErr(err)

                fmt.Printf("recieved %s from %v\n", value.Payload, remoteaddr)

                go sendResponse(serverConn, remoteaddr)
        }
}
