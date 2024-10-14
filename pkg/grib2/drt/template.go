package drt

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/scorix/grib-go/pkg/grib2/drt/datapacking"
	"github.com/scorix/grib-go/pkg/grib2/drt/definition"
	gridpoint "github.com/scorix/grib-go/pkg/grib2/drt/grid_point"
)

type TemplateNumber uint16

const (
	GridPointDataSimplePacking                        TemplateNumber = 0
	MatrixValueAtGridPointSimplePacking               TemplateNumber = 1
	GridPointDataComplexPacking                       TemplateNumber = 2
	GridPointDataComplexPackingAndSpatialDifferencing TemplateNumber = 3
	GridPointDataIEEEFloatingPointData                TemplateNumber = 4
	// 5-39 Reserved
	GridPointDataJPEG2000CodeStreamFormat TemplateNumber = 40
	GridPointDataPNG                      TemplateNumber = 41
	GridPointDataCCSDS                    TemplateNumber = 42
	// 43-49 Reserved
	SpectralDataSimplePacking  TemplateNumber = 50
	SpectralDataComplexPacking TemplateNumber = 51
	// 52 Reserved
	SpectralDataComplexPackinForLimitedAreaModels TemplateNumber = 53
	// 54-60 Reserved
	GridPointDataSimplePackingWithLogarithmPreProcessing TemplateNumber = 61
	// 62-199 Reserved
	RunLengthPackingWithLevelValues TemplateNumber = 200
	// 201-49151 Reserved
	// 49152-65534 Reserved For Local Use
	TemplateNumberMissing TemplateNumber = 255
)

type Template interface {
	GetNumVals() int
	Definition() any
	ReadAllData(r datapacking.BitReader) ([]float64, error)
}

func ReadTemplate(r datapacking.BitReader, n TemplateNumber, numVals int) (Template, error) {
	switch n {
	case GridPointDataSimplePacking:
		var tplDef definition.SimplePacking

		if err := binary.Read(r, binary.BigEndian, &tplDef); err != nil {
			return nil, err
		}

		return gridpoint.NewSimplePacking(tplDef, numVals), nil

	case GridPointDataComplexPacking:
		var tplDef definition.ComplexPacking

		if err := binary.Read(r, binary.BigEndian, &tplDef); err != nil {
			return nil, err
		}

		return gridpoint.NewComplexPacking(tplDef, numVals), nil

	case GridPointDataComplexPackingAndSpatialDifferencing:
		var tplDef definition.ComplexPackingAndSpatialDifferencing

		if err := binary.Read(r, binary.BigEndian, &tplDef); err != nil {
			return nil, err
		}

		return gridpoint.NewComplexPackingAndSpatialDifferencing(tplDef, numVals), nil

	}

	return nil, fmt.Errorf("data template not implemented: %d", n)
}

// TemplateMarshaler
type TemplateMarshaler struct {
	Template Template
}

type templateMarshaler struct {
	Number  TemplateNumber `json:"number"`
	Content string         `json:"content"`
	Vals    int            `json:"vals"`
}

func (tm TemplateMarshaler) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, tm.Template.Definition()); err != nil {
		return nil, err
	}

	var tplNum TemplateNumber
	switch tm.Template.(type) {
	case *gridpoint.SimplePacking:
		tplNum = GridPointDataSimplePacking
	case *gridpoint.ComplexPacking:
		tplNum = GridPointDataComplexPacking
	case *gridpoint.ComplexPackingAndSpatialDifferencing:
		tplNum = GridPointDataComplexPackingAndSpatialDifferencing
	}

	return json.Marshal(templateMarshaler{
		Number:  tplNum,
		Content: hex.EncodeToString(buf.Bytes()),
		Vals:    tm.Template.GetNumVals(),
	})
}

func (tm *TemplateMarshaler) UnmarshalJSON(data []byte) error {
	var t templateMarshaler

	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	decoded, err := hex.DecodeString(t.Content)
	if err != nil {
		return err
	}

	switch t.Number {
	case GridPointDataSimplePacking:
		var tplDef definition.SimplePacking

		if err := binary.Read(bytes.NewBuffer(decoded), binary.BigEndian, &tplDef); err != nil {
			return err
		}

		tm.Template = gridpoint.NewSimplePacking(tplDef, t.Vals)
		return nil

	case GridPointDataComplexPacking:
		var tplDef definition.ComplexPacking

		if err := binary.Read(bytes.NewBuffer(decoded), binary.BigEndian, &tplDef); err != nil {
			return err
		}

		tm.Template = gridpoint.NewComplexPacking(tplDef, t.Vals)
		return nil

	case GridPointDataComplexPackingAndSpatialDifferencing:
		var tplDef definition.ComplexPackingAndSpatialDifferencing

		if err := binary.Read(bytes.NewBuffer(decoded), binary.BigEndian, &tplDef); err != nil {
			return err
		}

		tm.Template = gridpoint.NewComplexPackingAndSpatialDifferencing(tplDef, t.Vals)
		return nil
	}

	return fmt.Errorf("data template not implemented: %d", t.Number)
}
