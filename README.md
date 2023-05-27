# [Prometheus](https://github.com/prometheus/prometheus) exporter framework

It helps you to easily build an exporter so that you only need to focus on the metrics themselves and not the exporter.

Different from the Prometheus official exporter, exporter framework use [Gin](https://github.com/gin-gonic/gin) as HTTP server and use [Logrus](https://github.com/sirupsen/logrus) as logger.

## Usage

First of all, register exporter:

```go
exporter.Register("name", "common_namespace", "Description.", ":{PORT}", logrusLogger)
```

Second, create some collectors:

```go
type myCollector struct {...}

func (c myCollector) Update(ch chan<- prometheus.Metric) error {...}
```

Then, create construct functions and register them in their `init` function:

```go
func newMyCollector(namespace string, logger *logrus.Entry) (exporter.Collector, error) {...}

func init() {
    exporter.RegisterCollector("collector_name", exporter.DefaultEnabled, newMyCollector)
}
```

Finally, run exporter:

```go
exporter.Run()
```

## Example

See [`_example`](https://github.com/rea1shane/exporter/tree/main/_example) directory.

## Tips

### environment variables

Gin sets itself using environment variables, view [Gin documentation](https://github.com/gin-gonic/gin#documentation) for more information.

> ATTENTION
>
> You must set `defaultAddress` to `""` during register exporter for gin to be able to read the environment variable `PORT`. View the `resolveAddress` function in the `exporter.go` for more information.
