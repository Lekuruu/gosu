# gosu

something something osu beatmap parser that is very cool and yeah this is a wip as you can probably tell but this will be really useful one day shoutout to the guy who made danser which i can just borrow the code from

## Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/Lekuruu/gosu/pkg/beatmaps"
)

func main() {
	file, err := os.Open("map.osu")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	beatmap, err := beatmaps.ParseFromReader(file)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s - %s mapped by %s\n", beatmap.Artist, beatmap.Title, beatmap.Creator)
}
```

## License

This repository is licensed under GPL-3.0. See [LICENSE](LICENSE).
Third-party attribution for gosu-pp is retained in [NOTICE](NOTICE).
