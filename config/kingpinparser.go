package config

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"strconv"
	"strings"
)

var flagStringMap = make(map[string]*string)

var flagBoolMap = make(map[string]*bool)

var flagFloatMap = make(map[string]*float64)

var flagIntMap = make(map[string]*int)


func String(name string, help string, defaultValue string) *string {
	if isExecutable() {
		return kingpin.Flag(name, help).Default(defaultValue).String()
	}
	// Else we need to overload and exclusively pull from the config file
	stringVal := new(string)
	stringVal = &defaultValue
	flagStringMap[name] = stringVal
	return stringVal
}


func StringNoDefault(name string, help string) *string {
	if isExecutable() {
		return kingpin.Flag(name, help).String()
	}
	// Else we need to overload and exclusively pull from the config file
	stringVal := new(string)
	flagStringMap[name] = stringVal
	return stringVal
}


func Bool(name string, help string) *bool {
	// If this is running as a separate executable instead of a library then return kingpin
	if isExecutable() {
		return kingpin.Flag(name, help).Bool()
	}
	// Else we need to overload and exclusively pull from the config file
	boolVal := new(bool)
	flagBoolMap[name] = boolVal
	return boolVal
}

func Float64(name string, help string, defaultValue string) *float64 {
	// If this is running as a separate executable instead of a library then return kingpin
	if isExecutable() {
		return kingpin.Flag(name, help).Default(defaultValue).Float64()
	}
	// Else we need to overload and exclusively pull from the config file
	floatVal := new(float64)
	flagFloatMap[name] = floatVal
	return floatVal
}


func Int(name string, help string, defaultValue string) *int {
	// If this is running as a separate executable instead of a library then return kingpin
	if isExecutable() {
		return kingpin.Flag(name, help).Default(defaultValue).Int()
	}
	// Else we need to overload and exclusively pull from the config file
	intVal := new(int)
	defInt, _ := strconv.Atoi(defaultValue)
	flagIntMap[name] = &defInt
	return intVal
}



func isExecutable() bool {
	// If this is running as a separate executable instead of a library then return kingpin
	// TODO would love a better check here
	return strings.Contains(os.Args[0],"exporter")
}


func LoadConfig(configYaml string) {
	values, _ := NewResolverFromFragment(configYaml)
	for flagName, flagValue := range values {
		if value, exist := flagStringMap[flagName]; exist {
			*value = flagValue
		}
	}
}

