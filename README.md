# Exporter

Fork from [`node_exporter`](https://github.com/prometheus/node_exporter) revision [`cf8c6891cc610e54f70383addd4bb6079f0add35`](https://github.com/prometheus/node_exporter/tree/cf8c6891cc610e54f70383addd4bb6079f0add35).

## Usage

### PProf statistics

Add follow to enable PProf statistics:

```go
package main

import (
	_ "net/http/pprof"
)
```

See [prometheus/exporter-toolkit#196](https://github.com/prometheus/exporter-toolkit/pull/196) for more information.
