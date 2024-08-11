//go:build windows

package ohwm

import (
	"errors"
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/prometheus-community/windows_exporter/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
)

const Name = "ohwm"

type Config struct{}

var ConfigDefaults = Config{}

// A collector is a Prometheus collector for CMI Sensor metrics
type collector struct {
	logger log.Logger

	Value *prometheus.Desc
	Min   *prometheus.Desc
	Max   *prometheus.Desc
}

func New(logger log.Logger, _ *Config) types.Collector {
	c := &collector{}
	c.SetLogger(logger)
	return c
}

func NewWithFlags(_ *kingpin.Application) types.Collector {
	return &collector{}
}

func (c *collector) GetName() string {
	return Name
}

func (c *collector) SetLogger(logger log.Logger) {
	c.logger = log.With(logger, "collector", Name)
}

func (c *collector) GetPerfCounter() ([]string, error) {
	return []string{}, nil
}

func (c *collector) Build() error {
	c.Value = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "value"), "(Value)", []string{
			"name", "identifier", "sensor_type", "parent", "index",
		}, nil)
	c.Min = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "min"), "(Min)", []string{
			"name", "identifier", "sensor_type", "parent", "index",
		}, nil)
	c.Max = prometheus.NewDesc(
		prometheus.BuildFQName(types.Namespace, Name, "max"), "(Max)", []string{
			"name", "identifier", "sensor_type", "parent", "index",
		}, nil)

	if err := ole.CoInitialize(1); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	return nil
}

// Collect sends the metric values for each metric
// to the provided prometheus Metric channel.
func (c *collector) Collect(_ *types.ScrapeContext, ch chan<- prometheus.Metric) error {
	if err := c.collect(ch); err != nil {
		_ = level.Error(c.logger).Log("msg", fmt.Sprintf("failed collecting ohwm metrics: %s", err.Error()))
		return err
	}
	return nil
}

type Sensor struct {
	Name       string
	Identifier string
	SensorType string
	Parent     string
	Value      float32
	Min        float32
	Max        float32
	Index      int32
}

func (c *collector) collect(ch chan<- prometheus.Metric) error {
	sensors, err := collectSensors(c.logger)
	if err != nil {
		_ = level.Error(c.logger).Log("msg", "failed to get sensor data",
			"err", err.Error())

		return nil
	}

	for i := range sensors {
		ch <- prometheus.MustNewConstMetric(
			c.Value,
			prometheus.GaugeValue,
			float64(sensors[i].Value),
			sensors[i].Name, sensors[i].Identifier, sensors[i].SensorType,
			sensors[i].Parent, strconv.Itoa(int(sensors[i].Index)),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Min,
			prometheus.GaugeValue,
			float64(sensors[i].Value),
			sensors[i].Name, sensors[i].Identifier, sensors[i].SensorType,
			sensors[i].Parent, strconv.Itoa(int(sensors[i].Index)),
		)

		ch <- prometheus.MustNewConstMetric(
			c.Max,
			prometheus.GaugeValue,
			float64(sensors[i].Value),
			sensors[i].Name, sensors[i].Identifier, sensors[i].SensorType,
			sensors[i].Parent, strconv.Itoa(int(sensors[i].Index)),
		)
	}

	return nil
}

func collectSensors(logger log.Logger) ([]*Sensor, error) {
	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return nil, fmt.Errorf("failed to create SWbem Locator: %w", err)
	}

	defer unknown.Release()

	wmi, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, fmt.Errorf("failed to query IID IDispatch interface: %w", err)
	}

	defer wmi.Release()

	serviceRaw, err := oleutil.CallMethod(wmi, "ConnectServer", ".", `root/OpenHardwareMonitor`)
	if err != nil {
		return nil, fmt.Errorf("failed to connect server: %w", err)
	}

	service := serviceRaw.ToIDispatch()
	if service == nil {
		return nil, errors.New("failed to call ConnectServer method")
	}

	defer service.Release()

	resultRaw, err := oleutil.CallMethod(service, "ExecQuery", `SELECT * FROM Sensor`)
	if err != nil {
		return nil, fmt.Errorf("failed to list all sensors in the service: %w", err)
	}

	result := resultRaw.ToIDispatch()
	if result == nil {
		return nil, errors.New("failed to list all sensors in the service")
	}

	defer result.Release()

	countVar, err := oleutil.GetProperty(result, "Count")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the total number of properties in the result: %w", err)
	}

	count := int(countVar.Val)

	sensors := make([]*Sensor, 0, count)

	for i := 0; i < count; i++ {
		itemRaw, iterErr := oleutil.CallMethod(result, "ItemIndex", i)
		if iterErr != nil {
			return nil, fmt.Errorf("failed to fetch item from index: %w", err)
		}

		item := itemRaw.ToIDispatch()
		if item == nil {
			return nil, errors.New("failed to fetch item from index")
		}

		sensor, iterErr := toSensor(item)

		if iterErr != nil {
			var name string
			if sensor != nil {
				name = sensor.Name
			}

			_ = level.Warn(logger).Log("msg", "failed to get sensor data",
				"name", name, "err", iterErr.Error())

			continue
		}

		sensors = append(sensors, sensor)
	}

	return sensors, nil
}

func toSensor(item *ole.IDispatch) (*Sensor, error) {
	sensor := &Sensor{}

	name, iterErr := oleutil.GetProperty(item, "Name")
	if iterErr != nil {
		return nil, fmt.Errorf("failed to get sensor's name: %w", iterErr)
	}

	sensor.Name = name.ToString()

	id, iterErr := oleutil.GetProperty(item, "Identifier")
	if iterErr != nil {
		return nil, fmt.Errorf("failed to get sensor's identifier: %w", iterErr)
	}

	sensor.Identifier = id.ToString()

	typ, iterErr := oleutil.GetProperty(item, "SensorType")
	if iterErr != nil {
		return nil, fmt.Errorf("failed to get sensor's type: %w", iterErr)
	}

	sensor.SensorType = typ.ToString()

	parent, iterErr := oleutil.GetProperty(item, "Parent")
	if iterErr != nil {
		return nil, fmt.Errorf("failed to get sensor's parent: %w", iterErr)
	}

	sensor.Parent = parent.ToString()

	valueRaw, iterErr := oleutil.GetProperty(item, "Value")
	if iterErr != nil {
		return nil, fmt.Errorf("failed to get sensor's value: %w", iterErr)
	}

	value, ok := valueRaw.Value().(float32)
	if !ok {
		return nil, fmt.Errorf("failed to convert sensor's value: %w", iterErr)
	}

	sensor.Value = value

	minValueRaw, iterErr := oleutil.GetProperty(item, "Min")
	if iterErr != nil {
		return nil, fmt.Errorf("failed to get sensor's min value: %w", iterErr)
	}

	minValue, ok := minValueRaw.Value().(float32)
	if !ok {
		return nil, fmt.Errorf("failed to convert sensor's min value: %w", iterErr)
	}

	sensor.Min = minValue

	maxValueRaw, iterErr := oleutil.GetProperty(item, "Max")
	if iterErr != nil {
		return nil, fmt.Errorf("failed to get sensor's max value: %w", iterErr)
	}

	maxValue, ok := maxValueRaw.Value().(float32)
	if !ok {
		return nil, fmt.Errorf("failed to convert sensor's max value: %w", iterErr)
	}

	sensor.Max = maxValue

	indexRaw, iterErr := oleutil.GetProperty(item, "Index")
	if iterErr != nil {
		return nil, fmt.Errorf("failed to get sensor's index: %w", iterErr)
	}

	index, ok := indexRaw.Value().(int32)
	if !ok {
		return nil, fmt.Errorf("failed to convert sensor's index: %w", iterErr)
	}

	sensor.Index = index

	return sensor, nil
}
