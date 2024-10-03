//go:build windows

package types

const (
	DefaultCollectors            = "cpu,cs,memory,logical_disk,physical_disk,net,os,service,system"
	DefaultCollectorsPlaceholder = "[defaults]"
	Namespace                    = "windows"
)
