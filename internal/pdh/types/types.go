package types

type Collector interface {
	Collect(dst any) error
	Close()
}
