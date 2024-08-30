//go:build windows

package types

const (
	DefaultCollectors            = "cpu,logical_disk,physical_disk,net,os,service,system"
	DefaultCollectorsPlaceholder = "[defaults]"
	Namespace                    = "windows"
)
