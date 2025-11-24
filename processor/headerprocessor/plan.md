# Header Processor Implementation Plan

## Overview

The Header Processor is a component that extracts HTTP headers from the request context and sets them as attributes on telemetry data (traces, metrics, logs). This processor enables users to capture header information for observability purposes, such as tracking request metadata, user agents, correlation IDs, or custom headers.

## Core Functionality

### Header Extraction
- Extract headers from the request context using the OpenTelemetry Collector's client metadata
- Support extraction of one or more specific headers by name
- Support extraction of all headers with optional filtering patterns
- Handle multiple values for a single header (as per HTTP specification)

### Attribute Setting
- Set extracted header values as attributes on telemetry data
- Support optional prefixing of attribute names
- Handle multiple header values by joining them with a configurable separator (default: `;`)
- Preserve original header names or allow custom attribute naming

## Configuration Schema

```yaml
processors:
  headers:
    # List of header extraction configurations
    headers:
      - name: "user-agent"           # Header name to extract
        attribute: "http.user_agent" # Target attribute name (optional, defaults to header name)
        prefix: "header."            # Optional prefix for attribute name
        
      - name: "x-correlation-id"
        attribute: "correlation_id"
        
      - name: "authorization"
        attribute: "auth_header"
        prefix: "request."
        
    # Global settings
    prefix: "http.header."           # Global prefix applied to all headers (optional)
    separator: ";"                   # Separator for multiple header values (default: ";")
    include_all: false               # Extract all headers (default: false)
    exclude_patterns:                # Regex patterns for headers to exclude when include_all is true
      - "^authorization$"
      - "^cookie$"
```

## Implementation Architecture

### 1. Configuration Structure
```go
type Config struct {
    Headers         []HeaderConfig `mapstructure:"headers"`
    GlobalPrefix    string         `mapstructure:"prefix"`
    Separator       string         `mapstructure:"separator"`
    IncludeAll      bool           `mapstructure:"include_all"`
    ExcludePatterns []string       `mapstructure:"exclude_patterns"`
}

type HeaderConfig struct {
    Name      string `mapstructure:"name"`
    Attribute string `mapstructure:"attribute"`
    Prefix    string `mapstructure:"prefix"`
}
```

### 2. Core Processing Logic

#### Context Header Extraction
- Use `client.FromContext(ctx)` to access request metadata
- Extract headers from `client.Metadata` using case-insensitive lookup
- Handle multiple values per header according to HTTP specification

#### Attribute Setting
- Apply prefixes in order of precedence: header-specific prefix > global prefix
- Generate attribute names: `[prefix][attribute_name || header_name]`
- Join multiple header values using the configured separator
- Set attributes on spans, metrics, and log records

### 3. Processing Flow
1. **Context Validation**: Verify that client metadata is available in the context
2. **Header Extraction**: Extract configured headers or all headers (if include_all is true)
3. **Filtering**: Apply exclude patterns when include_all is enabled
4. **Attribute Generation**: Create attribute key-value pairs with proper prefixing
5. **Telemetry Enhancement**: Set attributes on traces, metrics, and logs

## Key Features

### Multi-Value Header Support
- Handle headers with multiple values (e.g., `Accept: text/html, application/json`)
- Join values using configurable separator (default: `;`)
- Preserve all values to maintain complete header information

### Flexible Naming
- Support custom attribute names different from header names
- Apply prefixes at global or per-header level
- Use header name as default attribute name if not specified

### Security Considerations
- Provide exclude patterns to prevent extraction of sensitive headers
- Default exclusions for common sensitive headers (authorization, cookie)
- Case-insensitive header matching for robustness

## Use Cases

### 1. Request Tracing
```yaml
processors:
  headers:
    headers:
      - name: "x-trace-id"
        attribute: "trace.external_id"
      - name: "x-request-id"
        attribute: "request.id"
```

### 2. User Agent Analysis
```yaml
processors:
  headers:
    prefix: "http."
    headers:
      - name: "user-agent"
        attribute: "user_agent"
      - name: "accept"
        attribute: "accept"
```

### 3. Custom Header Extraction
```yaml
processors:
  headers:
    headers:
      - name: "x-tenant-id"
        attribute: "tenant.id"
        prefix: "custom."
      - name: "x-api-version"
        attribute: "api.version"
        prefix: "custom."
```

### 4. Comprehensive Header Capture
```yaml
processors:
  headers:
    include_all: true
    prefix: "http.header."
    exclude_patterns:
      - "^authorization$"
      - "^cookie$"
      - "^x-forwarded-.*"
```

## Implementation Steps

### Phase 1: Core Infrastructure
1. Define configuration structures and validation
2. Implement factory and component lifecycle
3. Create basic header extraction from context
4. Add unit tests for configuration and basic functionality

### Phase 2: Processing Logic
1. Implement header-to-attribute conversion
2. Add support for prefixes and custom attribute names
3. Handle multiple header values with configurable separator
4. Add processing for traces, metrics, and logs

### Phase 3: Advanced Features
1. Implement include_all functionality
2. Add exclude pattern filtering
3. Add case-insensitive header matching
4. Comprehensive integration testing

### Phase 4: Documentation and Examples
1. Create comprehensive README with examples
2. Add configuration documentation
3. Create example configurations for common use cases
4. Add performance benchmarks

## Dependencies

- `go.opentelemetry.io/collector/client` - For accessing request metadata
- `go.opentelemetry.io/collector/pdata` - For telemetry data manipulation
- `go.opentelemetry.io/collector/processor` - For processor interface
- Standard Go libraries for regex pattern matching and string manipulation

## Testing Strategy

### Unit Tests
- Configuration validation and parsing
- Header extraction logic
- Attribute name generation with prefixes
- Multiple value handling

### Integration Tests
- End-to-end processing with real telemetry data
- Context metadata extraction
- Performance testing with various header configurations

### Edge Cases
- Missing headers
- Empty header values
- Headers with special characters
- Large numbers of headers
- Invalid regex patterns in exclude_patterns

## Performance Considerations

- Minimize regex compilation by caching compiled patterns
- Efficient string operations for prefix application
- Lazy evaluation of include_all to avoid unnecessary processing
- Memory-efficient handling of multiple header values

## Security and Privacy

- Default exclusion of sensitive headers (authorization, cookie)
- Configurable exclude patterns for custom sensitive headers
- Case-insensitive matching to prevent bypass attempts
- Documentation of security best practices

## Future Enhancements

- Support for header value transformation (e.g., base64 decode)
- Conditional extraction based on header values
- Integration with authentication extensions for enhanced security
- Support for custom header value parsers
- Metrics for monitoring header extraction performance