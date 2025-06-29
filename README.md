# demux

[![CI](https://github.com/floatdrop/demux/actions/workflows/ci.yaml/badge.svg)](https://github.com/floatdrop/demux/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/floatdrop/demux)](https://goreportcard.com/report/github.com/floatdrop/demux)
[![Go Reference](https://pkg.go.dev/badge/github.com/floatdrop/demux.svg)](https://pkg.go.dev/github.com/floatdrop/demux)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

`demux` is a lightweight Go package that provides flexible and generic demultiplexing (fan-out) utilities for channels. It allows you to route items from a single input channel to multiple output channels based on dynamic or static keys.

## Features

- **Dynamic Demuxing**: Automatically spawn goroutines for each unique key, with dedicated channels.
- **Static Demuxing**: Route messages to pre-defined channels based on their keys.
- **Generic**: Uses Go generics for maximum flexibility.

## Installation

```bash
go get github.com/floatdrop/demux
```

## Usage

### Dynamic Demuxing

`Dynamic` demuxes messages from an input channel to a dynamically created set of output channels, one per key. Each unique key launches a dedicated goroutine running `consumeFunc`.

```golang
demux.Dynamic(input, func(msg MyType) string {
    return msg.UserID // or any key
}, func(key string, ch <-chan MyType) {
    for msg := range ch { // start consuming messages with same UserID
        fmt.Printf("Consumer for %s got: %+v\n", key, msg)
    }
})
```

### Static Demuxing

`Static` demuxes messages based on a key and routes them to pre-defined channels in a map.

```golang
channels := map[string]chan MyType{
    "alpha": make(chan MyType),
    "beta":  make(chan MyType),
}

go demux.Static(input, func(msg MyType) string {
    return msg.Group
}, channels)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
