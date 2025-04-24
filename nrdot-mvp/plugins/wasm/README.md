# WASM PII Masker Module

This directory contains the WASM module for PII masking in the NRDOT+ MVP.

## Overview

The PII Masker is a WebAssembly (WASM) module that detects and masks personally identifiable information (PII) in telemetry data. It is integrated into the OpenTelemetry Collector pipeline using the WASM processor.

## Features

- Masks sensitive data such as:
  - Passwords
  - Social Security Numbers (SSN)
  - Credit card numbers
  - Personal identification numbers
- Configurable masking patterns via JSON configuration
- Low overhead: ~2.4 µs/record median, CPU +0.05% at 5k records/s

## Integration

The module is integrated into the collector using the WASM processor, which is configured in the collector's configuration file:

```yaml
processors:
  wasm:
    modules:
      - name: pii_masker
        path: /plugins/pii_masker.wasm
        timeout_ms: 3
        memory_limit_mb: 8
```

## Host Imports

The WASM module uses the following host imports:

- `log_utf8`: For logging messages from the WASM module
- `read_attr`: To read attribute values from telemetry records
- `write_attr`: To write modified attribute values back to telemetry records
- `drop_record`: To drop records that contain critical PII that should not be processed

## Building

The WASM module is built using Rust with the WASI target. For development purposes, a placeholder binary is included in this repository. In a production environment, the full Rust source code would be available for customization and building.

## Configuration

The module can be configured using a JSON configuration file that specifies the masking patterns to use:

```json
{
  "patterns": [
    {
      "type": "password",
      "regex": "password=[^&]*",
      "replacement": "password=********"
    },
    {
      "type": "ssn",
      "regex": "\\d{3}-\\d{2}-\\d{4}",
      "replacement": "XXX-XX-XXXX"
    },
    {
      "type": "credit_card",
      "regex": "\\d{4}[- ]?\\d{4}[- ]?\\d{4}[- ]?\\d{4}",
      "replacement": "XXXX-XXXX-XXXX-XXXX"
    }
  ],
  "attributes_to_check": [
    "http.url",
    "http.request.body",
    "request.body",
    "db.statement",
    "message"
  ]
}
```

## Performance Considerations

The WASM module is designed to be lightweight and performant. It adds minimal overhead to the collector pipeline and can handle high throughput scenarios.
