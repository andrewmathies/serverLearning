package main

import (
        "fmt"
        "net"
        "time"
	"strconv"
	"bufio"
)
//

func main() {
	/*
	localAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")

	if err != nil {
		fmt.Printf("some error %v\n", err)
		return
	}
	*/
	serverAddr, err := net.ResolveUDPAddr("udp", "13.58.214.125:10001")

	if err != nil {
		fmt.Printf("some error %v\n", err)
		return
	}

        conn, err := net.DialUDP("udp", nil, serverAddr)

        if err != nil {
                fmt.Printf("some error %v\n", err)
                return
        }

	defer conn.Close()
	fmt.Printf("initialized\n")

	i := 0
	for {
		msg := strconv.Itoa(i)
		i++
		buf := []byte(msg)
		_, err := conn.Write(buf)
		
		if err != nil {
			fmt.Printf(msg, err)
			return
		}
		
		readBuf := make([]byte, 1024)
		placeholder, err := bufio.NewReader(conn).Read(readBuf)

		if err != nil {
			fmt.Printf("error reading! ", err)
			fmt.Printf("plus this! ", placeholder)
			return
		} else {
			fmt.Printf("%s\n\n", readBuf)
		}

		time.Sleep(time.Second * 1)
	}
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
