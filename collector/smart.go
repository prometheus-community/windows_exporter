// returns data points from MSStorageDriver_ATAPISmartData class

package collector

import (
	"log"

	"github.com/StackExchange/wmi"
	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	Factories["smart"] = NewSMARTCollector
}

// A SMARTCollector is a Prometheus collector for WMI metrics
type SMARTCollector struct {
	SelfTestStatus *prometheus.Desc
	TotalTime      *prometheus.Desc
	Capability     *prometheus.Desc
}

// NewSMARTCollector ...
func NewSMARTCollector() (Collector, error) {
	const subsystem = "smart"

	return &SMARTCollector{
		SelfTestStatus: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "selftest_status"),
			"The self test status code (SMART.SelfTestStatus)",
			[]string{"volume"},
			nil,
		),
		TotalTime: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "total_time"),
			"Total time used (SMART.TotalTime)",
			[]string{"volume"},
			nil,
		),
		Capability: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, subsystem, "capability"),
			"Smart capability (SMART.SmartCapability)",
			[]string{"volume"},
			nil,
		),
	}, nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *SMARTCollector) Collect(ch chan<- prometheus.Metric) error {
	if desc, err := c.collect(ch); err != nil {
		log.Println("[ERROR] failed collecting smart metrics:", desc, err)
		return err
	}
	return nil
}

type MSStorageDriver_ATAPISmartData struct {
	InstanceName    string
	Active          bool
	SelfTestStatus  uint64
	TotalTime       uint64
	SmartCapability uint64
	VendorSpecific  []uint8 // TODO parse this when https://github.com/StackExchange/wmi/pull/30 is merged
}

func (c *SMARTCollector) collect(ch chan<- prometheus.Metric) (*prometheus.Desc, error) {
	var dst []MSStorageDriver_ATAPISmartData
	if err := wmi.QueryNamespace(wmi.CreateQuery(&dst, ""), &dst, `root\wmi`); err != nil {
		return nil, err
	}

	poharg := "hour" // XXX add as command line arg, either "hour" (default), "min" or "sec"

	for _, disk := range dst {
		if !disk.Active {
			// exclude non-active disks
			continue
		}
		volume := disk.InstanceName
		ch <- prometheus.MustNewConstMetric(
			c.SelfTestStatus,
			prometheus.GaugeValue,
			float64(disk.SelfTestStatus),
			volume,
		)

		ch <- prometheus.MustNewConstMetric(
			c.TotalTime,
			prometheus.GaugeValue,
			float64(disk.TotalTime),
			volume,
		)

		ch <- prometheus.MustNewConstMetric(
			c.Capability,
			prometheus.GaugeValue,
			float64(disk.SmartCapability),
			volume,
		)

		rawreaderror := 0
		attrcriterror := 0
		hddattrcriterror := 0
		seekerror := 0
		dmacrcerror := 0
		softreaderror := 0
		programfailcount := 0
		erasefailcount := 0
		programfailcount2 := 0
		erasefailcount2 := 0

		spinavg := 0
		reallocatedsectors := 0
		spinretry := 0
		reserveblocks := 0
		endtoend := 0
		commandtimeout := 0

		lifetimeremain := 0
		lbawrite := 0

		reallocationevent := 0
		pendingsectors := 0
		uncorrectablesectors := 0

		temperature := 0
		tempmax := 0
		tempmin := 0

		poh := 0.

		for i := 0; i < len(disk.VendorSpecific); i += 11 {
			v := disk.VendorSpecific[i]

			if v == 0 || v == 16 { // field is 0 or 16? (only first row uses 16)
				if len(disk.VendorSpecific) < i+1 {
					log.Println("error parsing vendor specific data at", i, ":", disk.VendorSpecific)
					break
				}
				v = disk.VendorSpecific[i+1]
				log.Println("smart ", v)
				if v != 0 {
					log.Println("unexpected smart ", v)
					continue
				}

				i3 := disk.VendorSpecific[i+3]   // smart id
				i6 := disk.VendorSpecific[i+6]   // actual normalized data
				i7 := disk.VendorSpecific[i+7]   // worst normalized data
				i8 := disk.VendorSpecific[i+8]   // raw value as decimal
				i9 := disk.VendorSpecific[i+9]   // raw value as decimal
				i10 := disk.VendorSpecific[i+10] // raw value as decimal
				i11 := disk.VendorSpecific[i+11]
				i12 := disk.VendorSpecific[i+12]

				vendec := 0
				// attributes may have different ways of calculation
				switch i3 {
				case 4, 9, 193, 195, 200, 225, 241, 242, 246:
					// for those attributes where values up to 65k is not enough
					vendec = int(i12)*(16^8) + int(i11)*(16^6) + int(i10)*(16^4) + int(i9)*(16^2) + int(i8)
				case 194:
					// temperature is using only one field
					vendec = int(i8)
				default:
					// some attributes like id3 are using only 2 fields, other fields may display average or other things
					vendec = int(i9)*(16^2) + int(i8)
				}

				switch i3 {
				case 1:
					// set alarm if needed
					// some vendors use high raw values here on a new disc i.e. seagate
					// fujitsu is using only 2 fields
					rawreaderror = vendec
					if i6 <= 50 || i7 <= 50 {
						attrcriterror = attrcriterror + 1
					}
				case 3:
					// stores in only 2 fields, the other 2 are for average, the last one is unknown
					spinavg = int(i11)*(16^2) + int(i10)
					if i6 <= 50 || i7 <= 50 {
						attrcriterror = attrcriterror + 1
					}
				case 5:
					// Count of reallocated sectors. When the hard drive finds a read/write/verification error,
					// it marks that sector as "reallocated" and transfers data to a special reserved area
					// (spare area). a brand new disc has already reallocated sectors which are not shown, so
					// this value shouldnt really not increase because also the reserved area has a very
					// limited amount of space. fujitsu uses other fields for something else (hidden remaps?),
					// should be 0 anyway - ssd use higher values and indicate as failed flash memory blocks
					// on ssd this value increase as it ages
					reallocatedsectors = vendec
					if reallocatedsectors > 0 {
						hddattrcriterror = hddattrcriterror + 1
					}
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 7:
					// fujitsu seems to use less fields here
					// The raw value has different structure for different vendors and is often not meaningful as a decimal number.
					seekerror = vendec
					if i6 <= 60 || i7 <= 60 {
						attrcriterror = attrcriterror + 1
					}
				case 9:
					// some vendors use minutes or even seconds
					if poharg == "min" {
						poh = float64(vendec) / 60
					} else if poharg == "sec" {
						poh = float64(vendec) / 3600
					} else {
						poh = float64(vendec)
					}
				case 10:
					// Count of retry of spin start attempts. This attribute stores a total count of the spin
					// start attempts to reach the fully operational speed (under the condition that the first
					// attempt was unsuccessful). An increase of this attribute value is a sign of problems in
					// the hard disk mechanical subsystem.
					spinretry = vendec
					if spinretry > 0 {
						attrcriterror = attrcriterror + 1
					}
				case 170:
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
					reserveblocks = vendec
				case 171:
					// (Kingston)Counts the number of flash program failures. This Attribute returns the total
					// number of Flash program operation failures since the drive was deployed.
					// This attribute is identical to attribute 181.
					programfailcount = vendec
					if programfailcount > 0 {
						hddattrcriterror = hddattrcriterror + 1
					}
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 172:
					// (Kingston)Counts the number of flash erase failures. This Attribute returns the total
					// number of Flash erase operation failures since the drive was deployed.
					// This Attribute is identical to Attribute 182.
					erasefailcount = vendec
					if erasefailcount > 0 {
						hddattrcriterror = hddattrcriterror + 1
					}
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 173:
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 177:
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 179:
					// ssd reserved blocks shows remaining reserve blocks in percent
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 180:
					// reserved blocks
					reserveblocks = vendec
				case 181:
					// program fail count
					programfailcount2 = vendec
					if programfailcount2 > 0 {
						hddattrcriterror = hddattrcriterror + 1
					}
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 182:
					// "Pre-Fail" Attribute used at least in Samsung devices.
					erasefailcount2 = vendec
					if erasefailcount2 > 0 {
						hddattrcriterror = hddattrcriterror + 1
					}
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 183:
					// runtime bad block
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 184:
					// This attribute is a part of Hewlett-Packard's SMART IV technology, as well as part of
					// other vendors' IO Error Detection and Correction schemas, and it contains a count of
					// parity errors which occur in the data path to the media via the drive's cache RAM
					endtoend = vendec
					if endtoend > 0 {
						attrcriterror = attrcriterror + 1
					}
					if i6 <= 50 || i7 <= 50 {
						attrcriterror = attrcriterror + 1
					}
				case 188:
					// The count of aborted operations due to HDD timeout. Normally this attribute value should
					// be equal to zero and if the value is far above zero, then most likely there will be some
					// serious problems with power supply or an oxidized data cable. seen high raw values on
					// seagate discs in smartctl with normal thresholds, maybe only 2 fields are used
					commandtimeout = vendec
					if commandtimeout > 0 {
						attrcriterror = attrcriterror + 1
					}
				case 194:
					// temperature stores value only in one field
					temperature = vendec
					tempmax = int(i12)
					tempmin = int(i10)
				case 196:
					// critical, fujitsu uses other fields for something else, so dont use all fields together
					// many crucial m500 use 16 as raw value. ssd have increasing values over time
					reallocationevent = vendec
					if reallocationevent > 0 {
						hddattrcriterror = hddattrcriterror + 1
					}
				case 197:
					// critical value
					pendingsectors = vendec
					if pendingsectors > 0 {
						attrcriterror = attrcriterror + 1
					}
				case 198:
					// critical value
					uncorrectablesectors = vendec
					if uncorrectablesectors > 0 {
						attrcriterror = attrcriterror + 1
					}
				case 199:
					// mostly cable problems that should not happen
					dmacrcerror = vendec
					if dmacrcerror > 0 {
						attrcriterror = attrcriterror + 1
					}
				case 200:
					// the count of errors found when writing a sector.
					// The higher the value,the worse the disk's mechanical condition is.
					// uses more than 2 fields
					if i6 <= 99 || i7 <= 99 {
						attrcriterror = attrcriterror + 1
					}
				case 201:
					// Count of off-track errors.
					softreaderror = vendec
					if softreaderror > 0 {
						attrcriterror = attrcriterror + 1
					}
				case 202:
					// lifetime remaining in % on crucial ssd
					lifetimeremain = int(i6)
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 225:
					lbawrite = vendec
				case 226:
					// media war, value is remaining life in percent
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 230:
					// drive life protection status kingston
					if i7 <= 90 {
						attrcriterror = attrcriterror + 1
					}
				case 231:
					// Indicates the approximate SSD life left, in terms of program/erase cycles
					// or Flash blocks currently available for use
					lifetimeremain = int(i6)
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 232:
					// Available reserved space SSD
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 233:
					// ssd wearout indicator
					lifetimeremain = int(i6)
					if i6 <= 10 || i7 <= 10 {
						attrcriterror = attrcriterror + 1
					}
				case 241:
					// Total count of LBAs written
					lbawrite = vendec
				case 246:
					// Total count of LBAs written
					lbawrite = vendec
				}

			}
		}

	}

	return nil, nil
}
