package main

import (
    "os"
    "net"
    "bufio"
    "strconv"
    "fmt"
    "strings"
)

type ClientInfo struct {
    add bool
    connection net.Conn
    id string
}

// channel used for assign cliet id
var idAssignmentChan = make(chan string)
// map use to keep track of connection
var clientIdToStream = make(map[string] net.Conn)
// channel used to keep the connnt queue for the server
var contentQueue = make(chan string)
// channel used to keep track of whether add/delete a connection client in the map
var clientChnnel = make(chan ClientInfo)

/*
send the content to the conentQueue
can be called by all the client connection
*/
func SendToQueue(content string) {
    contentQueue <- content
}

/*
parse the content send by each client
pack the content,
with the format: object/clent : content
*/
func EncodeContent(content string, client_id string) {
    split_content := strings.Split(content, ":")
    command := strings.Trim(split_content[0], " ")
    contentInfo := strings.Trim(strings.Join(split_content[1:], ":"), " ")
    switch command {
    case "whoami":
        SendToQueue(client_id + ": chitter: " + client_id + "\r\n")
    case "all":
        SendToQueue("all: " + client_id + ": " + contentInfo)
    case "close":
        SendToQueue(content)
    default:
        SendToQueue(command + ": " + client_id + ": " + contentInfo)
    }
}

/*
gotutine to keep track of the content send by each client inside the contentQuest channel
*/
func SteamListener() {
    for {
        select {
        case content := <- contentQueue:
            DecodeContent(content)
        }
    }
}

/*
decode the content and execute the command sent by each client, only can be called by listener goroutine
*/
func DecodeContent(content string) {
    split_content := strings.Split(content, ":")
    command := strings.Trim(split_content[0], " ")
    contentInfo := strings.Trim(strings.Join(split_content[1:], ":"), " ")
    switch command {
    case "all":
        for id := range clientIdToStream {
            clientIdToStream[id].Write([]byte(string(contentInfo)))
        }
    case "close":
        // delete(clientIdToStream, contentInfo)
        DeleteConnetionRequest(contentInfo)
    default:
        if streamOfId, exist := clientIdToStream[command]; exist {
            streamOfId.Write([]byte(string(contentInfo)))
        }
    }
}

/*
each connection has a own goroutine to keep track of the message
*/
func HandleConnectionWithId(conn net.Conn, client_id string) {
    b := bufio.NewReader(conn)
    for {
        line, err := b.ReadBytes('\n')
        if err != nil {
            conn.Close()
            EncodeContent(string("close: " + client_id), client_id)
            break
        }
        EncodeContent(string(line), client_id)
    }
}

/*
assign id to connection
*/
func IdManager() {
    var i uint64
    for i = 0;  ; i++ {
        idAssignmentChan <- strconv.FormatUint(i, 10)
    }
}

/*
Decode info in the client information
*/
func DecodeClientInfo(clientInfo ClientInfo) {
    switch clientInfo.add {
    case true :
        clientIdToStream[clientInfo.id] = clientInfo.connection
    default:
        delete(clientIdToStream, clientInfo.id)
    }
}


/*
keep track of whether add or delete client in the map
*/
func ClientListener() {
    for {
        select {
        case clientInfo := <- clientChnnel:
            DecodeClientInfo(clientInfo)
        }
    }
}

func AddConnectionRequest(id string, stream net.Conn) {
    var addRequest = ClientInfo{add: true, connection: stream, id: id }
    clientChnnel <- addRequest
}

func DeleteConnetionRequest(id string) {
    var deleteRequest = ClientInfo {add: false, connection: nil, id: id}
    clientChnnel <- deleteRequest
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
    go ClientListener()

    fmt.Println("Listening on port", os.Args[1])
    for{
        conn, _ := server.Accept()
        // assign id to each connection
        // also, record the id - connection in the map
        client_id := <- idAssignmentChan
        AddConnectionRequest(client_id, conn)
        // clientIdToStream[client_id] = conn
        go HandleConnectionWithId(conn, client_id)
    }
}
