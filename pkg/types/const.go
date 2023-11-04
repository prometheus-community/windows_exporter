//go:build windows

package types

const (
	DefaultCollectors            = "cpu,cs,logical_disk,physical_disk,net,os,service,system,textfile"
	DefaultCollectorsPlaceholder = "[defaults]"
	Namespace                    = "windows"
)
