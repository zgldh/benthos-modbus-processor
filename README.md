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
        bytes_per_address: 2      # Optional, default 2
        data_length:              # Bytes length not Address length.
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
              raw_type: "Int16"   # Int16, Uint16, 
              big_endian: true    # Optional, default true.
            properties:
              value_type: "Float"
              scale: 0.1
          -
            name: "ThermostatH"
            attributes:
              starting_address: 0x11
              raw_type: "Int16"
              big_endian: true    # Optional, default true.
            properties:
              value_type: "Float"
              scale: 0.1
          -
            name: "AlarmMode"
            attributes:
              starting_address: 0x12
              raw_type: "Int16"
            properties:
              value_type: "Int"
          -
            name: "Temperature"
            attributes:
              starting_address: 0x13
              raw_type: "Int16"
              big_endian: true    # Optional, default true.
            properties:
              value_type: "Float"
              scale: 0.1
```


```json
Meta Data
{
  "modbus_crc_checked": true,
  "modbus_bytes_length": 123,
  "modbus_slave_id": 1
}

Payload
{
  "ThermostatL": 123.345,
  "ThermostatH": 123.345,
  "AlarmMode": 123.345,
  "Temperature": 123.345,
}
```