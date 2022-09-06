package processor

import (
	"context"
	"errors"
	"fmt"

	"encoding/json"

	"github.com/benthosdev/benthos/v4/public/service"
)

type RawType string

const (
	Int8    RawType = "Int8"
	Int16   RawType = "Int16"
	Int32   RawType = "Int32"
	Float32 RawType = "Float32"
	Float64 RawType = "Float64"
)

type ValueType string

const (
	Int   RawType = "Int"
	Float RawType = "Float"
)

type ConfigDataLength struct {
	StartingAddress int
	NumBytes        int
	BigEndian       bool
}

type ConfigCrc16 struct {
	Enabled   bool
	BigEndian bool
}

type ConfigFieldAttributes struct {
	StartingAddress int
	RawType         RawType //  "Int8", "Int16", "Int32", "Float32", "Float64"
	BigEndian       bool
}

type ConfigFieldProperties struct {
	ValueType ValueType // "Int", "Float"
	Scale     float64
}

type ConfigField struct {
	Name       string
	Attributes ConfigFieldAttributes
	Properties ConfigFieldProperties
}

type ModbusProcessorConfig struct {
	BytesPerAddress int
	DataLength      ConfigDataLength
	Crc16           ConfigCrc16
	Fields          []ConfigField
}
type ModbusProcessor struct {
	config *ModbusProcessorConfig
	logger *service.Logger
}

func newModbusProcessor(conf *service.ParsedConfig, logger *service.Logger) (*ModbusProcessor, error) {
	bytes_per_address, err := conf.FieldInt("bytes_per_address")
	if err != nil {
		bytes_per_address = 2
	}

	data_length__starting_address, err := conf.FieldInt("data_length", "starting_address")
	if err != nil {
		data_length__starting_address = 2
	}
	data_length__num_bytes, err := conf.FieldInt("data_length", "num_bytes")
	if err != nil {
		data_length__num_bytes = 2
	}
	data_length__big_endian, err := conf.FieldBool("data_length", "big_endian")
	if err != nil {
		data_length__big_endian = true
	}

	crc16__enabled, err := conf.FieldBool("crc16", "enabled")
	if err != nil {
		crc16__enabled = true
	}
	crc16__big_endian, err := conf.FieldBool("crc16", "big_endian")
	if err != nil {
		crc16__big_endian = true
	}

	configFields, err := conf.FieldObjectList("fields")
	if err != nil {
		return nil, errors.New("fields are required")
	}
	fields := make([]ConfigField, len(configFields))

	for index, configField := range configFields {
		name, err := configField.FieldString("name")
		if err != nil {
			return nil, errors.New("field name is required")
		}
		attributes_starting_address, err := configField.FieldInt("attributes", "starting_address")
		if err != nil {
			return nil, fmt.Errorf("field '%v' attributes.starting_address is required", name)
		}
		attributes_raw_type, err := configField.FieldString("attributes", "raw_type")
		if err != nil {
			return nil, fmt.Errorf("field '%v' attributes.raw_type is required", name)
		}
		attributes_big_endian, err := configField.FieldBool("attributes", "big_endian")
		if err != nil {
			attributes_big_endian = true
		}

		properties_value_type, err := configField.FieldString("properties", "value_type")
		if err != nil {
			return nil, fmt.Errorf("field '%v' properties.value_type is required", name)
		}
		properties_scale, err := configField.FieldFloat("properties", "scale")
		if err != nil {
			properties_scale = 1
		}

		fields[index] = ConfigField{
			Name: name,
			Attributes: ConfigFieldAttributes{
				StartingAddress: attributes_starting_address,
				RawType:         RawType(attributes_raw_type),
				BigEndian:       attributes_big_endian,
			},
			Properties: ConfigFieldProperties{
				ValueType: ValueType(properties_value_type),
				Scale:     properties_scale,
			},
		}
	}

	return &ModbusProcessor{
		config: &ModbusProcessorConfig{
			BytesPerAddress: bytes_per_address,
			DataLength: ConfigDataLength{
				StartingAddress: data_length__starting_address,
				NumBytes:        data_length__num_bytes,
				BigEndian:       data_length__big_endian,
			},
			Crc16: ConfigCrc16{
				Enabled:   crc16__enabled,
				BigEndian: crc16__big_endian,
			},
			Fields: fields,
		},
		logger: logger,
	}, nil
}

func init() {
	modbusConfigSpec := service.NewConfigSpec().
		Summary("Creates a processor for Modbus data.").
		Field(service.NewIntField("bytes_per_address").Advanced().Default(2)).
		Field(service.NewObjectField("data_length",
			service.NewIntField("starting_address").Default(0x02),
			service.NewIntField("num_bytes").Default(2),
			service.NewBoolField("big_endian").Default(true),
		).Advanced().Default(ConfigDataLength{2, 2, true})).
		Field(service.NewObjectField("crc16",
			service.NewBoolField("enabled").Default(true),
			service.NewBoolField("big_endian").Default(true),
		).Advanced().Default(ConfigCrc16{true, true})).
		Field(service.NewObjectListField("fields",
			service.NewStringField("name"),
			service.NewObjectField("attributes",
				service.NewIntField("starting_address"),
				service.NewStringEnumField("raw_type", "Int8", "Int16", "Int32", "Float32", "Float64"), // TODO
				service.NewBoolField("big_endian").Default(true),
			),
			service.NewObjectField("properties",
				service.NewStringEnumField("value_type", "Int", "Float"), // TODO
				service.NewFloatField("scale").Default(1),
			),
		))

	constructor := func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
		return newModbusProcessor(conf, mgr.Logger())
	}

	err := service.RegisterProcessor("modbus", modbusConfigSpec, constructor)
	if err != nil {
		panic(err)
	}
}

func (r *ModbusProcessor) Process(ctx context.Context, m *service.Message) (service.MessageBatch, error) {
	// bytesContent, err := m.AsBytes()
	// if err != nil {
	// 	return nil, err
	// }
	newBytes, err := json.Marshal(*r.config)
	if err != nil {
		panic(err)
	}

	m.SetBytes(newBytes)
	return []*service.Message{m}, nil
}

func (r *ModbusProcessor) Close(ctx context.Context) error {
	return nil
}
