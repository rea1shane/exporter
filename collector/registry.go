package collector

import (
	"fmt"

	"github.com/alecthomas/kingpin/v2"
	"github.com/sirupsen/logrus"
)

const (
	DefaultEnabled  = true
	DefaultDisabled = false
)

var (
	factories        = make(map[string]func(namespace string, logger *logrus.Entry) (Collector, error)) // factories records all collector's construction method
	collectorState   = make(map[string]*bool)                                                           // collectorState records all collector's default state (enabled or disabled)
	forcedCollectors = map[string]bool{}                                                                // forcedCollectors will record collectors that have explicitly declared state
)

func RegisterCollector(collector string, isDefaultEnabled bool, factory func(namespace string, logger *logrus.Entry) (Collector, error)) {
	var helpDefaultState string
	if isDefaultEnabled {
		helpDefaultState = "enabled"
	} else {
		helpDefaultState = "disabled"
	}

	flagName := fmt.Sprintf("collector.%s", collector)
	flagHelp := fmt.Sprintf("Enable the %s collector (default: %s).", collector, helpDefaultState)
	defaultValue := fmt.Sprintf("%v", isDefaultEnabled)

	flag := kingpin.Flag(flagName, flagHelp).Default(defaultValue).Action(collectorFlagAction(collector)).Bool()
	collectorState[collector] = flag

	factories[collector] = factory
}

// collectorFlagAction generates a new action function for the given collector
// to track whether it has been explicitly enabled or disabled from the command line.
// A new action function is needed for each collector flag because the ParseContext
// does not contain information about which flag called the action.
// See: https://github.com/alecthomas/kingpin/issues/294
func collectorFlagAction(collector string) func(ctx *kingpin.ParseContext) error {
	return func(ctx *kingpin.ParseContext) error {
		forcedCollectors[collector] = true
		return nil
	}
}

// DisableDefaultCollectors sets the collector state to false for all collectors which
// have not been explicitly enabled on the command line.
func DisableDefaultCollectors() {
	for c := range collectorState {
		if _, ok := forcedCollectors[c]; !ok {
			*collectorState[c] = false
		}
	}
}
