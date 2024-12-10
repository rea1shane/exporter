# [Prometheus](https://github.com/prometheus/prometheus) exporter framework

It helps you to easily build an exporter so that you only need to focus on the metrics themselves and not the exporter. Fork from [`node_exporter`](https://github.com/prometheus/node_exporter).

You can manage collectors like node exporter. For example:

- [Enable & Disable collectors](https://github.com/prometheus/node_exporter/?tab=readme-ov-file#collectors)
- [Include & Exclude flags](https://github.com/prometheus/node_exporter/?tab=readme-ov-file#include--exclude-flags)
- [Filtering enabled collectors](https://github.com/prometheus/node_exporter/?tab=readme-ov-file#filtering-enabled-collectors)
- And more...

Run exporter with `-h` to see all available configuration flags.

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
