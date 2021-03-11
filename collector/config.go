package collector

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configMap         = make(map[string]Config)
	configInstanceMap = make(map[string]*ConfigInstance)
)

// Used to hold the metadata about a configuration option
type Config struct {
	Name     string
	HelpText string
	Default  string
}

// Used to hold the actual values for a configuration option
type ConfigInstance struct {
	Value      string
	IsValueSet bool
	Config
}

// This interface is used when a Collector needs to have configuration. This code should support multiple collectors of
// the same type which means we cannot use the global var based configuration.
type ConfigurableCollector interface {
	ApplyConfig(map[string]*ConfigInstance)
}

func addConfig(config []Config) {
	for _, v := range config {
		ci := &ConfigInstance{
			Value:  "",
			Config: v,
		}
		configInstanceMap[v.Name] = ci
		configMap[v.Name] = v
	}
}

func ApplyKingpinConfig(app *kingpin.Application) map[string]*ConfigInstance {
	// associate each kingpin var with a var in the instance map
	for _, v := range configInstanceMap {
		app.Flag(v.Name, v.HelpText).Default(v.Default).Action(setExists).StringVar(&v.Value)
	}
	return configInstanceMap
}

// This exists mostly to support the Bool parameter
// Its incredibly hard to determine --boolValue with default false, since kingpin returns "" and since I am treating
// everything has a string instead of using the BoolVar
// In this case when the parameter is parsed we set the exists flag
func setExists(ctx *kingpin.ParseContext) error {
	for _, v := range ctx.Elements {
		name := ""
		if c, ok := v.Clause.(*kingpin.CmdClause); ok {
			name = c.Model().Name
		} else if c, ok := v.Clause.(*kingpin.FlagClause); ok {
			name = c.Model().Name
		} else if c, ok := v.Clause.(*kingpin.ArgClause); ok {
			name = c.Model().Name
		} else {
			continue
		}
		// There are some high level configurations that dont apply to collectors that
		// dont exist in the config instance map
		value, exists := configInstanceMap[name]
		if exists == false {
			continue
		}
		value.IsValueSet = true
	}

	return nil
}

// Function used by collectors to get the value, returns the default if config has not been set
func getValueFromMap(m map[string]*ConfigInstance, key Config) string {
	if v, configExists := m[key.Name]; configExists {
		if v.IsValueSet {
			return v.Value
		}
		return v.Default
	}
	return ""
}
