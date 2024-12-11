# [Prometheus](https://github.com/prometheus/prometheus) exporter framework

Lets you create a powerful exporter in two minutes. See [Usage](#usage).

## Features

This framework is extracted from [`node_exporter`](https://github.com/prometheus/node_exporter). You can use all the features of `node_exporter` with this framework:

- [Enable & Disable collectors](https://github.com/prometheus/node_exporter/?tab=readme-ov-file#collectors)
- [Include & Exclude flags](https://github.com/prometheus/node_exporter/?tab=readme-ov-file#include--exclude-flags)
- [Filtering enabled collectors](https://github.com/prometheus/node_exporter/?tab=readme-ov-file#filtering-enabled-collectors)
- Useful metrics `collector_duration_seconds` and `collector_success` for collectors.
- ...

## Example

There is an example in [`_example`](https://github.com/rea1shane/exporter/tree/main/_example).

## Usage

Creating your export is very simple:

1. Create some collectors implement `github.com/rea1shane/exporter/collector/Collector` and call `github.com/rea1shane/exporter/collector/RegisterCollector` in their `init` function.
2. Call the `github.com/rea1shane/exporter/Run` function to start the exporter.

Same as `node_exporter`, the framework uses `log/slog` as the logger and `github.com/alecthomas/kingpin/v2` as the command line argument parser.

These tips will help you create a better exporter:

- `collector.ErrNoData` indicates the collector found no data to collect, but had no other error. If necessary, return it in the collector's `Update` method.
- `metric.TypedDesc` makes easier to create metrics.
- If you are not using `metric.TypedDesc` to build metrics, you can use `util.AnyToFloat64` to convert the data to `float64`.

### Optional features

#### PProf statistics

Add `_ "net/http/pprof"` in imports to enable PProf statistics:

```go
package main

import (
	_ "net/http/pprof"
)
```

See [prometheus/exporter-toolkit#196](https://github.com/prometheus/exporter-toolkit/pull/196) for more information.
