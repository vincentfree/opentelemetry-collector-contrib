# Baggage Processor

| Status        |                                      |
|---------------|--------------------------------------|
| Stability     | [development]: traces, metrics, logs |
| Distributions | [contrib]                            |

[development]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/component-stability.md#development
[beta]: https://github.com/open-telemetry/opentelemetry-collector#beta

[core]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol
[contrib]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
[k8s]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-k8s

The Baggage processor provides special handling for
the [W3C Baggage propagation format](https://www.w3.org/TR/baggage/). It enables extraction, injection, modification,
and filtering of baggage entries according to the W3C Baggage specification.

## Overview

W3C Baggage is a propagation format for user-defined properties in distributed contexts. It enables the transmission of
key-value pairs through requests, allowing systems to correlate events beyond request identification covered by trace
context specifications.

The Baggage processor supports the following operations:

- **Extract**: Extract baggage entries from the context and add them as attributes
- **Inject**: Inject attributes as baggage entries into the context
- **Update**: Update existing baggage entries with new values
- **Upsert**: Insert or update baggage entries
- **Delete**: Remove baggage entries from the context

## Configuration

The processor supports the following configuration options:

```yaml
baggage:
  # List of actions to perform on baggage entries (required)
  actions:
    - key: <baggage_key>
      action: <extract|inject|update|upsert|delete>
      # Additional action-specific configuration...

  # Prefix to add to attribute names when extracting baggage (optional)
  # Default: "baggage."
  attribute_prefix: "baggage."

  # Maximum size in bytes for the baggage header (optional)
  # Default: 8192 bytes (W3C specification default)
  max_baggage_size: 8192

  # Whether to drop invalid baggage entries instead of failing (optional)
  # Default: false
  drop_invalid_baggage: false
```

### Action Configuration

Each action supports the following fields:

#### Common Fields

- `key` (string, required): The baggage key to operate on
- `action` (string, required): The type of action to perform

#### Extract Action

Extracts baggage entries from the context and adds them as attributes.

```yaml
- key: user.id
  action: extract
  from_context: true                    # Required: must be true
  to_attribute: custom.user.id          # Optional: custom attribute name
```

#### Inject Action

Injects attributes as baggage entries into the context.

```yaml
- key: service.version
  action: inject
  value: "1.2.3"                        # Option 1: static value
  # OR
  from_attribute: service.version       # Option 2: from attribute
  properties: # Optional: W3C baggage properties
    priority: high
    source: config
```

#### Update Action

Updates existing baggage entries with new values (skips if key doesn't exist).

```yaml
- key: session.timeout
  action: update
  from_attribute: session.timeout_ms
  properties:
    unit: milliseconds
```

#### Upsert Action

Inserts or updates baggage entries (creates if key doesn't exist).

```yaml
- key: feature.flags
  action: upsert
  from_attribute: enabled_features
  properties:
    format: json
    version: v1
```

#### Delete Action

Removes baggage entries from the context.

```yaml
- key: temp.debug.info
  action: delete
```

## Use Cases

### 1. E-commerce Application

Extract customer context from baggage and inject request metadata:

```yaml
processors:
  baggage/ecommerce:
    attribute_prefix: "baggage."
    max_baggage_size: 8192
    actions:
      # Extract customer context
      - key: customer.id
        action: extract
        from_context: true
      - key: customer.tier
        action: extract
        from_context: true
        to_attribute: customer.membership.tier

      # Inject request context
      - key: request.id
        action: inject
        from_attribute: http.request.id
      - key: feature.flags
        action: inject
        from_attribute: feature.enabled_flags
        properties:
          format: json
          version: v1
```

### 2. Microservices Tracing

Handle distributed tracing context with baggage:

```yaml
processors:
  baggage/microservices:
    attribute_prefix: "trace.baggage."
    drop_invalid_baggage: true
    actions:
      # Extract distributed tracing context
      - key: trace.parent.service
        action: extract
        from_context: true
      - key: trace.correlation.id
        action: extract
        from_context: true
        to_attribute: correlation.id

      # Inject service metadata
      - key: service.instance.id
        action: inject
        from_attribute: service.instance.id
      - key: deployment.version
        action: inject
        from_attribute: deployment.version
        properties:
          environment: production
          region: us-east-1
```

### 3. Security and Compliance

Handle security context and compliance metadata:

```yaml
processors:
  baggage/security:
    attribute_prefix: "security.baggage."
    max_baggage_size: 4096
    drop_invalid_baggage: true
    actions:
      # Extract security context
      - key: auth.token.type
        action: extract
        from_context: true
      - key: auth.permissions
        action: extract
        from_context: true
        to_attribute: user.permissions

      # Inject compliance metadata
      - key: compliance.gdpr.consent
        action: inject
        from_attribute: gdpr.consent_status
        properties:
          version: "2.0"
          timestamp: auto

      # Remove sensitive debug information
      - key: debug.internal.state
        action: delete
```

## W3C Baggage Specification Compliance

This processor implements the [W3C Baggage specification](https://www.w3.org/TR/baggage/) with the following features:

- **Format Compliance**: Supports the standard baggage string format with key-value pairs and properties
- **Size Limits**: Configurable maximum baggage size (default 8192 bytes as per W3C spec)
- **Property Support**: Full support for baggage member properties
- **Error Handling**: Configurable behavior for invalid baggage entries
- **Encoding**: Proper handling of URL encoding for baggage keys and values

### Baggage String Format

The processor handles baggage strings in the W3C format:

```
key1=value1;property1=prop_value1;property2=prop_value2,key2=value2
```

### Properties

Baggage properties are metadata associated with baggage members:

```yaml
properties:
  priority: high        # Custom property
  source: processor     # Source identifier
  timestamp: auto       # Timestamp property
  format: json          # Data format hint
```

## Pipeline Configuration

The baggage processor can be used in traces, metrics, and logs pipelines:

```yaml
service:
  pipelines:
    traces:
      receivers: [ otlp ]
      processors: [ baggage/extract, batch ]
      exporters: [ otlp ]

    metrics:
      receivers: [ otlp ]
      processors: [ baggage/inject, batch ]
      exporters: [ otlp ]

    logs:
      receivers: [ otlp ]
      processors: [ baggage/security, batch ]
      exporters: [ otlp ]
```

## Performance Considerations

- **Baggage Size**: Large baggage can impact performance. Use `max_baggage_size` to limit overhead
- **Action Order**: Actions are processed in the order specified in the configuration
- **Error Handling**: Set `drop_invalid_baggage: true` for better resilience in production
- **Attribute Prefixes**: Use meaningful prefixes to avoid attribute name conflicts

## Troubleshooting

### Common Issues

1. **Invalid Baggage Format**: Enable `drop_invalid_baggage` to skip malformed entries
2. **Size Limits**: Adjust `max_baggage_size` based on your requirements
3. **Missing Attributes**: Ensure source attributes exist before injection actions
4. **Property Validation**: Check that property keys and values follow W3C specification

### Logging

The processor logs warnings and errors for:

- Invalid baggage members or properties
- Baggage size limit violations
- Missing source attributes for injection
- Action processing failures
