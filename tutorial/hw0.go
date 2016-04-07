package main

import (
    "os"
    "net"
    "bufio"
    "strconv"
    "fmt"
    "strings"
)

/*
new structure created for pass message inside the system
@id:
    "all": a message will be send to all the connected client
    default: assigned id for the connected id, send message to it/ delete it/ add it in the map
@add:
    "1": a new client come in, record the assigned channel to it in the global map to keep track of
    "-1": a client lost connection, delete the id - client channel in the map
    "0": a message will be sent inside the system
@channel:
    only valid when record the new connection client,
    record the channel assign to the client
@connect:
    only valid when passing message in the system
    content to be distributed to the specific id
*/
type ClientInfo struct {
    id string
    add string
    channel chan string
    content string
}

// channel used for assign cliet id
var idAssignmentChan = make(chan string)
// map use to keep track of channel created for each connection client
var clientIdToChannel = make(map[string] chan string)
// channel used to keep track of whether process the message sent by different client
var clientChnnel = make(chan ClientInfo)

/*
create goroutine for each connected client to listen for message send by self/others
*/
func ChannelListenerForClient(conn net.Conn, id string, private_channel chan string) {
	for content := range private_channel{
        conn.Write([]byte(string(content)))
	}
}

/*
called when distributed the content to all the connected client
*/
func BroadcastToAll(content string){
    for id := range clientIdToChannel {
        clientIdToChannel[id] <- content
    }
}

/*
send private message to a single connected client
if the client id does not existed, do nothing
*/
func TalkToSingle(id string, content string) {
    if channel_of_id, exist := clientIdToChannel[id]; exist {
        channel_of_id <- content
    }
}

/*
parse message by each client, to decide the command or where to send the message
*/
func ParseContent(content string, clientId string) {
    split_content := strings.Split(content, ":")

    var command, contentInfo string
    if len(split_content) < 2{
        // when there is no object assigned
        // consider it as send message to all
        command = "all"
        contentInfo = content
    } else {
        command = strings.Trim(split_content[0], " ")
        contentInfo = strings.Trim(strings.Join(split_content[1:], ":"), " ")
    }
    switch command {
    case "whoami":
        TalkConnectionRequest(clientId, "chitter: " + clientId + "\r\n")
    case "all":
        TalkConnectionRequest("all", clientId + ": " + contentInfo)
    default:
        TalkConnectionRequest(command, clientId + ": " + contentInfo)
    }
}

/*
each connection/client has own goroutine to keep track of message send by this client
*/
func HandleConnection(conn net.Conn, clientId string) {
    b := bufio.NewReader(conn)
    for {
        line, err := b.ReadBytes('\n')
        if err != nil {
            conn.Close()
            // if the connection is lost, delete its listening channel from global map
            DeleteConnetionRequest(clientId)
            break
        }
        ParseContent(string(line), clientId)
    }
}

/*
goroutine to continuesly assign id to connected client
*/
func IdManager() {
    var i uint64
    for i = 0;  ; i++ {
        idAssignmentChan <- strconv.FormatUint(i, 10)
    }
}

/*
called when a new message need to be passed between the clients
*/
func SendMessage(obj string, content string){
    switch obj {
    case "all":
        BroadcastToAll(content)
    default:
        TalkToSingle(obj, content)
    }
}

/*
Decode info in the client information
*/
func DecodeClientInfo(clientInfo ClientInfo) {
    switch clientInfo.add {
    case "1" :
        clientIdToChannel[clientInfo.id] = clientInfo.channel
    case "-1":
        delete(clientIdToChannel, clientInfo.id)
    default:
        SendMessage(clientInfo.id, clientInfo.content)
    }
}

/*
listen message in the channel with ClientInfo, call decode fuction to deal with the message
*/
func ClientListener() {
    for {
        select {
        case clientInfo := <- clientChnnel:
            DecodeClientInfo(clientInfo)
        }
    }
}

/*
called when new client connected and the channel assigned to it need to be recored in global map
*/
func AddConnectionRequest(id string, channel chan string) {
    var addRequest = ClientInfo{add: "1", channel: channel, id: id, content: ""}
    clientChnnel <- addRequest
}

/*
called when client lost connection and it's channel needs to be deleted from the global map
*/
func DeleteConnetionRequest(id string) {
    var deleteRequest = ClientInfo {add: "-1", channel: nil, id: id, content: ""}
    clientChnnel <- deleteRequest
}

/*
called when a message need to be passed between the clients
*/
func TalkConnectionRequest(id string, content string) {
    var talkRequest = ClientInfo {add: "0", channel: nil, id: id, content: content}
    clientChnnel <- talkRequest
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
    // start the goroutine to keep track of the message inside the clientChnnel
    go ClientListener()

    fmt.Println("Listening on port", os.Args[1])
    for{
        conn, _ := server.Accept()
        // assign id to the new connection
        clientId := <- idAssignmentChan
        // create new channel for each new connection
        var channelForId = make(chan string)
        // record the id and the channel for the new connection globally to be scheduled
        AddConnectionRequest(clientId, channelForId)
        // create goroutine to keep track of the connection coming info and the channel assigned to it
        go ChannelListenerForClient(conn, clientId, channelForId)
        go HandleConnection(conn, clientId)
    }
}
