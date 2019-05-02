package etwlogrus

import (
	"github.com/Microsoft/go-winio/internal/etw"
	"testing"
)

func fireEvent(t *testing.T, p *etw.Provider, name string, value interface{}) {
	if err := p.WriteEvent(
		name,
		nil,
		etw.WithFields(getFieldOpt("Field", value))); err != nil {

		t.Fatal(err)
	}
}

// The purpose of this test is to log lots of different field types, to test the
// logic that converts them to ETW. Because we don't have a way to
// programatically validate the ETW events, this test has two main purposes: (1)
// validate nothing causes a panic while logging (2) allow manual validation that
// the data is logged correctly (through a tool like WPA).
func TestFieldLogging(t *testing.T) {
	// Sample WPRP to collect this provider:
	//
	// <?xml version="1.0"?>
	// <WindowsPerformanceRecorder Version="1">
	//   <Profiles>
	//     <EventCollector Id="Collector" Name="MyCollector">
	//       <BufferSize Value="256"/>
	//       <Buffers Value="100"/>
	//     </EventCollector>
	//     <EventProvider Id="HookTest" Name="5e50de03-107c-5a83-74c6-998c4491e7e9"/>
	//     <Profile Id="Test.Verbose.File" Name="Test" Description="Test" LoggingMode="File" DetailLevel="Verbose">
	//       <Collectors>
	//         <EventCollectorId Value="Collector">
	//           <EventProviders>
	//             <EventProviderId Value="HookTest"/>
	//           </EventProviders>
	//         </EventCollectorId>
	//       </Collectors>
	//     </Profile>
	//   </Profiles>
	// </WindowsPerformanceRecorder>
	//
	// Start collection:
	// wpr -start HookTest.wprp -filemode
	//
	// Stop collection:
	// wpr -stop HookTest.etl
	p, err := etw.NewProvider("HookTest", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := p.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	fireEvent(t, p, "Bool", true)
	fireEvent(t, p, "BoolSlice", []bool{true, false, true})
	fireEvent(t, p, "EmptyBoolSlice", []bool{})
	fireEvent(t, p, "String", "teststring")
	fireEvent(t, p, "StringSlice", []string{"sstr1", "sstr2", "sstr3"})
	fireEvent(t, p, "EmptyStringSlice", []string{})
	fireEvent(t, p, "Int", int(1))
	fireEvent(t, p, "IntSlice", []int{2, 3, 4})
	fireEvent(t, p, "EmptyIntSlice", []int{})
	fireEvent(t, p, "Int8", int8(5))
	fireEvent(t, p, "Int8Slice", []int8{6, 7, 8})
	fireEvent(t, p, "EmptyInt8Slice", []int8{})
	fireEvent(t, p, "Int16", int16(9))
	fireEvent(t, p, "Int16Slice", []int16{10, 11, 12})
	fireEvent(t, p, "EmptyInt16Slice", []int16{})
	fireEvent(t, p, "Int32", int32(13))
	fireEvent(t, p, "Int32Slice", []int32{14, 15, 16})
	fireEvent(t, p, "EmptyInt32Slice", []int32{})
	fireEvent(t, p, "Int64", int64(17))
	fireEvent(t, p, "Int64Slice", []int64{18, 19, 20})
	fireEvent(t, p, "EmptyInt64Slice", []int64{})
	fireEvent(t, p, "Uint", uint(21))
	fireEvent(t, p, "UintSlice", []uint{22, 23, 24})
	fireEvent(t, p, "EmptyUintSlice", []uint{})
	fireEvent(t, p, "Uint8", uint8(25))
	fireEvent(t, p, "Uint8Slice", []uint8{26, 27, 28})
	fireEvent(t, p, "EmptyUint8Slice", []uint8{})
	fireEvent(t, p, "Uint16", uint16(29))
	fireEvent(t, p, "Uint16Slice", []uint16{30, 31, 32})
	fireEvent(t, p, "EmptyUint16Slice", []uint16{})
	fireEvent(t, p, "Uint32", uint32(33))
	fireEvent(t, p, "Uint32Slice", []uint32{34, 35, 36})
	fireEvent(t, p, "EmptyUint32Slice", []uint32{})
	fireEvent(t, p, "Uint64", uint64(37))
	fireEvent(t, p, "Uint64Slice", []uint64{38, 39, 40})
	fireEvent(t, p, "EmptyUint64Slice", []uint64{})
	fireEvent(t, p, "Uintptr", uintptr(41))
	fireEvent(t, p, "UintptrSlice", []uintptr{42, 43, 44})
	fireEvent(t, p, "EmptyUintptrSlice", []uintptr{})
	fireEvent(t, p, "Float32", float32(45.46))
	fireEvent(t, p, "Float32Slice", []float32{47.48, 49.50, 51.52})
	fireEvent(t, p, "EmptyFloat32Slice", []float32{})
	fireEvent(t, p, "Float64", float64(53.54))
	fireEvent(t, p, "Float64Slice", []float64{55.56, 57.58, 59.60})
	fireEvent(t, p, "EmptyFloat64Slice", []float64{})

	type struct1 struct {
		A    float32
		priv int
		B    []uint
	}
	type struct2 struct {
		A int
		B int
	}
	type struct3 struct {
		struct2
		A    int
		B    string
		priv string
		C    struct1
		D    uint16
	}
	// Unexported fields, and fields in embedded structs, should not log.
	fireEvent(t, p, "Struct", struct3{struct2{-1, -2}, 1, "2s", "-3s", struct1{3.4, -4, []uint{5, 6, 7}}, 8})
}
