package input

import (
	"github.com/benthosdev/benthos/v4/public/service"
)

type ConfigSpecDataLength struct {
	starting_address int
	num_bytes        int
	big_endian       bool
}

type ConfigSpecCrc16 struct {
	enabled    bool
	big_endian bool
}

var modbusConfigSpec = service.NewConfigSpec().
	Summary("Creates a processor for Modbus data.").
	Field(service.NewObjectField("data_length",
		service.NewIntField("starting_address").Default(0x02),
		service.NewIntField("num_bytes").Default(2),
		service.NewBoolField("big_endian").Default(true),
	).Advanced().Default(ConfigSpecDataLength{2, 2, true})).
	Field(service.NewObjectField("crc16",
		service.NewBoolField("enabled").Default(true),
		service.NewBoolField("big_endian").Default(true),
	).Advanced().Default(ConfigSpecCrc16{true, true})).
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

// TODO
func newModbusProcessor(conf *service.ParsedConfig) (service.Processor, error) {
	return nil, nil
}

func init() {
	err := service.RegisterProcessor(
		"modbus", modbusConfigSpec,
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
			return newModbusProcessor(conf)
		})
	if err != nil {
		panic(err)
	}
}
