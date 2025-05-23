package main

import (
    "fmt"
    "github.com/valorm/snapurl/internal/config"
)

func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", cfg)
}
