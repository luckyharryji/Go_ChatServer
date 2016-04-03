package main

import (
    "fmt"
    "os"
)

func main() {
    if len(os.Args) < 2{
        fmt.Fprintf(os.Stderr, "Usage: chitter <port-number>\n")
        os.Exit(1)
        return
    }
    port := os.Args[1]
    fmt.Println("I'll be using port:",port)
}
