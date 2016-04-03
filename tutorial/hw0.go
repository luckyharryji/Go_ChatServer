package main

import (
    "os"
    "net"
    "bufio"
    "strconv"
    "fmt"
)

var idAssignmentChan = make(chan string)
var client_id_to_stream = make(map[string] net.Conn)


func HandleConnection(conn net.Conn) {
    b := bufio.NewReader(conn)
    client_id := <-idAssignmentChan
    for {
        line, err := b.ReadBytes('\n')
        if err != nil {
            conn.Close()
            break
        }
        client_id_to_stream[client_id] = conn
        if client_id == "1" {
            client_id_to_stream["0"].Write([]byte(client_id + ": " +string(line)))
        } else{
            conn.Write([]byte(client_id + ": " +string(line)))
        }

    }
}

func IdManager() {
    var i uint64
    for i = 0;  ; i++ {
        idAssignmentChan <- strconv.FormatUint(i, 10)
    }
}

func main() {
    if len(os.Args) < 2{
        fmt.Fprintf(os.Stderr, "Usage: chitter <port-number>\n")
        os.Exit(1)
        return
    }
    port := os.Args[1]
    server, err := net.Listen("tcp", ":"+ port )
    if err != nil {
        fmt.Fprintln(os.Stderr, "Can't connect to port")
        os.Exit(1)
    }
    go IdManager()
    fmt.Println("Listening on port", os.Args[1])
    for{
        conn, _ := server.Accept()
        go HandleConnection(conn)
    }
}
