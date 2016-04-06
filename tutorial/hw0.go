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
// var client_id_to_stream = make(map[string] net.Conn)
var clientIdToChannel = make(map[string] chan string)
var clientIdToChannelTest = make(map[string] chan string)

var idQueue = make(chan string)


func Write(conn net.Conn, id string, private_channel chan string) {
	for content := range private_channel{
        conn.Write([]byte(string(content)))
	}
}

func WriteToAll(content string){
    for id := range clientIdToChannel {
        clientIdToChannel[id] <- content
    }
}

func TalkToSingle(id string, content string) {
    if channel_of_id, exist := clientIdToChannel[id]; exist {
        // fmt.Println(content + " " + id, " is talking to single")
        channel_of_id <- content
        // fmt.Println(content + " " + id, " send to single")
    }
}

func PutIdToQueue(client_id string) {
    idQueue <- client_id
}


func ParseContent(content string, client_id string) {
    fmt.Println(content, " is parsing")
    split_content := strings.Split(content, ":")
    command := strings.Trim(split_content[0], " ")
    contentInfo := strings.Trim(strings.Join(split_content[1:], ":"), " ")
    switch command {
    case "whoami":
        TalkToSingle(client_id, "chitter: " + client_id + "\r\n")
    case "all":
        WriteToAll(client_id + ": " + contentInfo)
    default:
        TalkToSingle(command, client_id + ": " + contentInfo)
    }
}

func HandleConnection(conn net.Conn, client_id string, private_channel chan string) {
    b := bufio.NewReader(conn)

    go Write(conn, client_id, private_channel)
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

// func CreateChannelForId(){
//     for id:= range idQueue{
//
//         // id := <- idQueue
//
//         fmt.Println(id, " is assigning")
//         var channelForId = make(chan string)
//         channelForId <- (id + " is ready")
//         clientIdToChannelTest[id] = channelForId
//     }
// }

func CreateChannelForId(){
    for {
        var channelForId = make(chan string)
        select {
        case id := <- idQueue:
            fmt.Println(id, " is assigning")
            clientIdToChannelTest[id] = channelForId
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

    // go CreateChannelForId()

    fmt.Println("Listening on port", os.Args[1])
    for{
        client_id := <- idAssignmentChan
        var channelForId = make(chan string)
        clientIdToChannel[client_id] = channelForId
        conn, _ := server.Accept()
        go HandleConnection(conn, client_id, channelForId)
    }
}
