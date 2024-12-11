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

1. Create some collectors implement `github.com/rea1shane/exporter/collector.Collector` and call `github.com/rea1shane/exporter/collector.RegisterCollector` in their `init` function.
2. Call the `github.com/rea1shane/exporter.Run` function to start the exporter.

Now, everything is done!

### Tips

- Same as `node_exporter`, the framework uses `log/slog` as the logger and `github.com/alecthomas/kingpin/v2` as the command line argument parser.
- `github.com/rea1shane/exporter/collector.ErrNoData` indicates the collector found no data to collect, but had no other error. If necessary, return it in the `github.com/rea1shane/exporter/collector.Collector`'s `Update` method.
- `github.com/rea1shane/exporter/metric.TypedDesc` makes easier to create metrics.
- If you are not using `github.com/rea1shane/exporter/metric.TypedDesc` to create metrics, you can use `github.com/rea1shane/exporter/util.AnyToFloat64` function to convert the data to `float64`.

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

## Development building and running

You can use [prometheus/promu](https://github.com/prometheus/promu) to build exporter that use this framework:

1. Copy [prometheus/prometheus/Makefile.common](https://github.com/prometheus/prometheus/blob/main/Makefile.common) to your repository root path.
2. Create `.promu.yml` like [prometheus/node_exporter/.promu.yml](https://github.com/prometheus/node_exporter/blob/master/.promu.yml) or [prometheus/blackbox_exporter/.promu.yml](https://github.com/prometheus/blackbox_exporter/blob/master/.promu.yml).
3. Create `Makefile` like [prometheus/node_exporter/Makefile](https://github.com/prometheus/node_exporter/blob/master/Makefile) or [prometheus/blackbox_exporter/Makefile](https://github.com/prometheus/blackbox_exporter/blob/master/Makefile).
4. Reference [`node_exporter`'s readme](https://github.com/prometheus/node_exporter?tab=readme-ov-file#development-building-and-running) or [`blackbox_exporter`'s readme (contains build with Docker)](https://github.com/prometheus/blackbox_exporter?tab=readme-ov-file#building-the-software) to complete the building and running.
