package main
import (
        "fmt"
        "net"
)

func sendResponse(conn *net.UDPConn, addr *net.UDPAddr) {
        _,err := conn.WriteToUDP([]byte("From server: Hello I got your message "), addr)

        if err != nil {
                fmt.Printf("Couldn't send err response %v", err)
        }
}

func main() {
        p := make([]byte, 2048)

        serverAddr, err := net.ResolveUDPAddr("udp", ":10001")

        if err != nil {
                fmt.Printf("Some error %v\n", err)
        }

        serverConn, err := net.ListenUDP("udp", serverAddr)

        if err != nil {
                fmt.Printf("Some error %v\n", err)
                return
        }
        
        defer serverConn.Close()
        fmt.Printf("Running, waiting for a response!\n")

        for {
                n, remoteaddr, err := serverConn.ReadFromUDP(p)
                fmt.Printf("recieved ", string(p[0:n]), " from ", remoteaddr, "\n\n")

                if err != nil {
                        fmt.Printf("some error %v", err)
                        continue
                }

                go sendResponse(serverConn, remoteaddr)
        }
}