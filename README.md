# [Prometheus](https://github.com/prometheus/prometheus) exporter framework

It helps you to easily build an exporter so that you only need to focus on the metrics themselves and not the exporter.

Exporter framework use [Kingpin](https://github.com/alecthomas/kingpin) as command line and flag parser, use [Gin](https://github.com/gin-gonic/gin) as HTTP server and use [Logrus](https://github.com/sirupsen/logrus) as logger.

## Usage

1.  Register exporter:

    ```go
    exporter.Register("name", "common_namespace", "Description.", ":{PORT}", logrusLogger)
    ```

1.  Create some collectors:

    ```go
    type myCollector struct {...}
    
    func (c myCollector) Update(ch chan<- prometheus.Metric) error {...}
    ```

1.  Create collectors' construct functions and register them in their `init` function:

    ```go
    func newMyCollector(namespace string, logger *logrus.Entry) (exporter.Collector, error) {...}
    
    func init() {
        exporter.RegisterCollector("collector_name", exporter.DefaultEnabled, newMyCollector)
    }
    ```

1.  Run exporter:

    ```go
    exporter.Run()
    ```

## Example

See [`_example`](https://github.com/rea1shane/exporter/tree/main/_example) directory. Run `cd _example && go run exporter.go -h` for more information.

## Tips

### environment variables

Gin sets itself using environment variables, view [Gin documentation](https://github.com/gin-gonic/gin#documentation) for more information.

> ATTENTION
>
> You must set `defaultAddress` to `""` during register exporter for gin to be able to read the environment variable `PORT`. View the `resolveAddress` function in the [`exporter.go`](https://github.com/rea1shane/exporter/blob/main/exporter.go) for more information.
