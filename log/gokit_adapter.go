package log

import (
	"github.com/go-kit/kit/log/level"
)

// Returns an adapter implementing the go-kit/kit/log.Logger interface on our
// logrus logger
func NewToolkitAdapter() *logAdapter {
	return &logAdapter{}
}

type logAdapter struct{}

func (*logAdapter) Log(keyvals ...interface{}) error {
	var lvl level.Value
	var msg string
	for i := 0; i < len(keyvals); i += 2 {
		switch keyvals[i] {
		case "level":
			tlvl, ok := keyvals[i+1].(level.Value)
			if !ok {
				Warnf("Could not cast level of type %T", keyvals[i+1])
			} else {
				lvl = tlvl
			}
		case "msg":
			msg = keyvals[i+1].(string)
		}
	}

	switch lvl {
	case level.ErrorValue():
		Errorln(msg)
	case level.WarnValue():
		Warnln(msg)
	case level.InfoValue():
		Infoln(msg)
	case level.DebugValue():
		Debugln(msg)
	default:
		Warnf("Unmatched log level: '%v' for message %q", lvl, msg)
	}

	return nil
}
