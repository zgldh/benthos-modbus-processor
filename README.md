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
        bytes_per_address: 2      # Optional, default 2. How many bytes per address for fields.
        data_length:              # Bytes length not Address length.
          byte_index: 0x02        # Optional, default 0x02. The index of the first byte for data length.
          bytes_num: 1            # Optional, default 1. How many bytes to define the data length. 1,2,4,8
          big_endian: true        # Optional, default true.
        crc_checking:
          enabled: true           # Optional, default true. The CRC matching value is the last 2, 4 or 8  bytes. It will take all bytes before the CRC matching value to get calculated CRC value.  It will throw exception if enabled and crc checking failed.
          type: MODBUS            # Optional, default MODBUS. Options: X25, CCITT, MODBUS, CRC16, XMODEM, XMODEM2, CRC32, IEEE, Castagnoli, CRC32C, Koopman, CRC64ISO, CRC64ECMA
          big_endian: false       # Optional, default true
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
  "modbus_slave_address": 1,
  "modbus_function": 3,
  "modbus_data_length": 123,
}

Payload
{
  "ThermostatL": 123.345,
  "ThermostatH": 123.345,
  "AlarmMode": 123.345,
  "Temperature": 123.345,
}
```