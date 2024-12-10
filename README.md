# [Prometheus](https://github.com/prometheus/prometheus) exporter framework

It helps you to easily build an exporter so that you only need to focus on the metrics themselves and not the exporter. Fork from [`node_exporter`](https://github.com/prometheus/node_exporter). Future updates to the `node_exporter` will also be merged into this repository.

You can manage collectors like node exporter. For example:

- [Enable & Disable collectors](https://github.com/prometheus/node_exporter/?tab=readme-ov-file#collectors)
- [Include & Exclude flags](https://github.com/prometheus/node_exporter/?tab=readme-ov-file#include--exclude-flags)
- [Filtering enabled collectors](https://github.com/prometheus/node_exporter/?tab=readme-ov-file#filtering-enabled-collectors)
- And more...

Run your exporter with `-h` to see all available configuration flags.

## Usage

Creating an export is very simple:

1. Create some collectors implement `collector/Collector` and call `collector/RegisterCollector` in their `init` function.
2. Call the `Run` function to start the exporter.

Tips:

- `collector.ErrNoData` indicates the collector found no data to collect, but had no other error. If necessary, return it in the collector's `Update` method.
- `metric.TypedDesc` makes easier to build metrics.
- If you are not using `metric.TypedDesc` to build metrics, you can use `util.AnyToFloat64` to convert the data to `float64`.
- Find more things in source code...

There is an example in [`_example`](https://github.com/rea1shane/exporter/tree/main/_example). It can help you get up to speed with the framework faster.

## Optional feature

### PProf statistics

Add follow to enable PProf statistics:

```go
package main

import (
	_ "net/http/pprof"
)
```

See [prometheus/exporter-toolkit#196](https://github.com/prometheus/exporter-toolkit/pull/196) for more information.
