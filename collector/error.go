package collector

import (
	"errors"
)

// ErrNoData indicates the collector found no data to collect, but had no other error.
var ErrNoData = errors.New("collector returned no data")

func isNoDataError(err error) bool {
	return err == ErrNoData
}
