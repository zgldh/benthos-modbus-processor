(Work In Progress)
=================
Benthos with Modbus Processor plugin
======================

## Build

```sh
go build
```

Alternatively build it as a Docker image with:

```sh
go mod vendor
docker build . -t benthos_modbus_processor
```

## Usage

```yaml
pipeline:
  processors:
    - modbus:
        data_length:
          starting_address: 0x02  # Optional, default 0x02
          num_bytes: 2            # Optional, default 2
          big_endian: true        # Optional, default true
        crc16:
          enabled: true           # Optional, default true. It will throw exception if enabled and crc16 checking failed.
          big_endian: true        # Optional, default true
        fields:
          -
            name: "ThermostatL"
            attributes:
              starting_address: 0x10
              raw_type: "Int16"
              big_endian: true    # Optional, default true.
            properties:
              value_type: "Float32"
              scale: "0.1"
          -
            name: "ThermostatH"
            attributes:
              starting_address: 0x12
              raw_type: "Int16"
              big_endian: true    # Optional, default true.
            properties:
              value_type: "Float32"
              scale: "0.1"
          -
            name: "AlarmMode"
            attributes:
              { starting_address: 0x14 }
            properties:
              value_type: "Int16"
          -
            name: "Temperature"
            attributes:
              starting_address: 0x16
              raw_type: "Int16"
              big_endian: true    # Optional, default true.
            properties:
              value_type: "Float32"
              scale: "0.1"
```