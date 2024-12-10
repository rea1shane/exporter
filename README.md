# Exporter

Fork from node exporter branch `master` commit `cf8c6891cc610e54f70383addd4bb6079f0add35`.

add follow to enable pprof:

```go
package main

import (
	_ "net/http/pprof"
)
```

See [Add pprof links to landing page by SuperQ · Pull Request #196 · prometheus/exporter-toolkit](https://github.com/prometheus/exporter-toolkit/pull/196)
