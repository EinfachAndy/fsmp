# Fast Fixed-Size Memory Poll (fsmp)

This is an implementation of the paper: computation_tools_2012_1_10_80006.pdf

## Features:
 - no golang garbage collection
 - no loops (fast access times)
 - no recursive functions
 - little initialization overhead
 - little memory footprint (few dozen bytes)
 - straightforward and trouble-free algorithm
 - no-fragmentation
 - control and organization of memory
 - usage of a spinlock for fast concurrent access

## Install

```
go get -u github.com/EinfachAndy/fsmp/
```

## Usage

```go
package main

import (
	"encoding/binary"

	"github.com/EinfachAndy/fsmp"
)

func main() {
	pool := fsmp.CreatePool(1, 8)
	b, err := pool.Allocate()
	if err == fsmp.ErrOutOfMemory {
		panic(err.Error())
	}
	value := uint64(1234)
	binary.LittleEndian.PutUint64(b, value)

	pool.DeAllocate(b)
}
```

## License

fsmp source code is available under the MIT [License](/LICENSE).
