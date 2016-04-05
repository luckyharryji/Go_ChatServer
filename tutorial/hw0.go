package main

import (
    "os"
    "net"
    "bufio"
    "strconv"
    "fmt"
    "strings"
)

var idAssignmentChan = make(chan string)
var clientIdToStream = make(map[string] net.Conn)
var contentQueue = make(chan string)

func SendToQueue(content string) {
    contentQueue <- content
}

func ParseContent(content string, client_id string) {
    split_content := strings.Split(content, ":")
    command := strings.Trim(split_content[0], " ")
    contentInfo := strings.Trim(strings.Join(split_content[1:], ":"), " ")
    switch command {
    case "whoami":
        SendToQueue(client_id + ": chitter: " + client_id + "\r\n")
    case "all":
        SendToQueue("all: " + client_id + ": " + contentInfo)
    default:
        SendToQueue(command + ": " + client_id + ": " + contentInfo)
    }
}

func SendContent(content string) {
    split_content := strings.Split(content, ":")
    command := strings.Trim(split_content[0], " ")
    contentInfo := strings.Trim(strings.Join(split_content[1:], ":"), " ")
    switch command {
    case "all":
        for id := range clientIdToStream {
            clientIdToStream[id].Write([]byte(string(contentInfo)))
        }
    default:
        clientIdToStream[command].Write([]byte(string(contentInfo)))
    }
}


func HandleConnectionWithId(conn net.Conn, client_id string) {
    b := bufio.NewReader(conn)
    for {
        line, err := b.ReadBytes('\n')
        if err != nil {
            conn.Close()
            break
        }
        ParseContent(string(line), client_id)
    }
}

func IdManager() {
    var i uint64
    for i = 0;  ; i++ {
        idAssignmentChan <- strconv.FormatUint(i, 10)
    }
}


func SteamListener() {
    for {
        select {
        case content := <- contentQueue:
            SendContent(content)
        }
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
    go SteamListener()

    fmt.Println("Listening on port", os.Args[1])
    for{
        conn, _ := server.Accept()
        client_id := <- idAssignmentChan
        clientIdToStream[client_id] = conn
        go HandleConnectionWithId(conn, client_id)
    }
}
