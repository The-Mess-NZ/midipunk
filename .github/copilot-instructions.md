# MidiPunk

## About the project
A compact, open source Raspberry Pi powered MIDI router written in Go.
It uses a TFT28 Touch Shield for RPi (v2.1), resistive touch screen for user interface.

## Hardware
Midipunk provides:
- 4x USB-A ports, for bi-directional USB-MIDI compliant devices
- 4x full-sized 5-pin MIDI DIN ports, which can be assigned as inputs, or outputs.

It runs on a Raspberry Pi 4B.

## Architecture Overview

### Core Components
- **Main Entry Point** (`midipunk.go`): Initializes configuration, creates router, starts USB detection and API server
- **Configuration** (`config/config.go`): YAML-based config loading with `MidiPunkConfig` struct containing ports and routes
- **MIDI Ports** (`midiport/`): Manages USB and DIN port configurations with `PortConfig` struct and device detection
- **Router** (`midirouter/`): Core routing logic with `Route`, `RouteInput`, `RouteOutput` structs for message distribution  
- **Logger** (`logger/`): Custom async logger with MIDI timing message filtering and buffered channels
- **API Server** (`api/`): REST endpoints for configuration management and WebSocket status/events

### Data Flow
1. Configuration loaded from `config.yaml` defines ports and routing rules
2. `PortConfig` objects created for each port (USB via ALSA client numbers, DIN via serial paths)
3. `Route` objects contain multiple `RouteInput` and `RouteOutput` combinations
4. `Router` singleton manages all routes, listening on inputs and forwarding messages to outputs
5. USB device hot-plugging detected via polling `/dev/snd` directory every 2 seconds
6. Messages flow through buffered channels: raw MIDI → wrapped with source → router → outputs

## Configuration Patterns

### YAML Structure
```yaml
logLevel: "none"  # none, error, info, debug, all (timing messages only at "all")
ports:
  - id: "1"              # USB: ALSA client number as string
    type: "USB"
    direction: "INPUT"
    label: "88 Keys In"   # Optional UI display name
  - id: "din1"
    type: "DIN"
    device: "/dev/ttyAMA2"  # Serial device path for DIN ports
    direction: "INPUT"
    label: "SQ-64 In"
routes:
  - inputs:
      - port_id: "1"
      - port_id: "din1"
    outputs:
      - port_id: "2"
        channel: 2         # Optional channel mapping (1-16)
      - port_id: "3" 
        channel: 2
```

### Port Configuration Patterns
- **USB Ports**: Use numeric string IDs matching ALSA client numbers (e.g., "1", "20")
- **DIN Ports**: Use descriptive IDs with `device` field pointing to serial paths like `/dev/ttyAMA2`
- **Direction**: Always uppercase "INPUT"/"OUTPUT" in YAML, converted via `PortDirectionFromString()`
- **Channels**: Default to channel 1 if not specified in route inputs/outputs
- **Labels**: Auto-generated from device names if not provided

## Coding Patterns

### Port Creation
```go
// USB port - ID matches ALSA client number
pc := midiport.NewUSBPortConfig("1", "", midiport.PortDirectionFromString("INPUT"), "USB Device")

// DIN port - requires device path
pc := midiport.NewDINPortConfig("din1", "/dev/ttyAMA2", midiport.PortDirectionFromString("INPUT"), "DIN Port")
```

### Route Construction
```go
inputs := []*midirouter.RouteInput{
    midirouter.NewRouteInput(1, usbPortConfig, "Input 1"),
}
outputs := []*midirouter.RouteOutput{
    midirouter.NewRouteOutput(2, dinPortConfig, "Output 1"), // Channel 2
}
route := midirouter.NewRoute(inputs, outputs, "Main Route")
```

### Message Flow Pattern
- Messages wrapped in `MessageWithSource` struct containing source port and channel info
- Router uses `shouldRouteMessage()` to match inputs to routes
- Channel mapping applied via `applyChannelChange()` for outputs
- USB sender functions cached in router's `senders` map to avoid reopening ports

### Error Handling
- Use `fmt.Errorf()` for wrapping errors with context
- Log errors with `logger.Error()` for async logging
- Return error tuples from functions that can fail
- Non-blocking channel sends with `select` to prevent deadlocks

## Build & Development

### Dependencies
- `gitlab.com/gomidi/midi/v2`: MIDI message handling and ALSA integration
- `go.bug.st/serial`: Serial communication for DIN ports at 38400 baud
- `gopkg.in/yaml.v3`: Configuration file parsing
- `github.com/gorilla/websocket`: WebSocket support for API

### Build Process
```bash
go build -o midipunk midipunk.go
```

### Runtime Requirements
- ALSA development tools (`aconnect` command) for USB MIDI device enumeration
- Access to `/dev/snd/` for USB device detection
- Serial port access for DIN MIDI (typically `/dev/ttyAMA2` on RPi)
- Go 1.24.2+ (as specified in go.mod)

## API Endpoints

### Configuration Management
- `GET /config/get`: Retrieve current YAML configuration
- `PUT /config/put`: Update configuration (accepts YAML body)

### Real-time Communication
- `GET /status/ws`: WebSocket for router status updates
- `GET /events/ws`: WebSocket for MIDI device connection events

## USB Device Detection

### Polling Mechanism
- Polls `/dev/snd/` directory every 2 seconds for changes
- Detects new/removed MIDI devices by comparing file listings
- Extracts device metadata from sysfs (`/sys/class/sound/` and USB device paths)
- Uses `aconnect -l` to map ALSA cards to client numbers
- Stores devices in global `USBMIDIDeviceRepo` map

### Device Repository Structure
- `USBMIDIDevice`: Stores vendor ID, product ID, serial, manufacturer, product name
- Indexed by device path (e.g., `/dev/snd/midiC1D0`)
- Includes ALSA client numbers for port creation

## Logging System

### Custom Logger Features
- Async logging via buffered channels (1000 message capacity)
- Special handling for MIDI timing messages (Clock, Start, Stop, etc.)
- Log levels: NONE, ERROR, INFO, DEBUG, ALL
- Timing messages only shown at ALL level to reduce noise
- Non-blocking sends prevent MIDI processing delays

### Usage Patterns
```go
logger.Info("General information")
logger.DebugMIDI("Regular MIDI message: %s", msg)
logger.DebugMIDITiming("Timing message: %s", msg)  // Only at ALL level
```

## MIDI Message Processing

### Input Processing
1. `PortConfig.StartListening()` creates goroutines for each input port
2. USB: Uses `midi.ListenTo()` with ALSA port numbers from gomidi library
3. DIN: Reads from serial port at 38400 baud, processes 3-byte MIDI messages
4. Raw messages sent to per-input channels, then wrapped with source info

### Routing Logic
- Router receives `MessageWithSource` structs in `handleMessages()` goroutine
- `shouldRouteMessage()` checks if message source/channel matches route input
- Channel-specific routing: extracts channel from MIDI message bytes (0x80-0xEF range)
- `applyChannelChange()` modifies message channel bits for outputs
- Sender functions cached per output port to avoid repeated port opening

### Current Limitations
- No MIDI message filtering beyond channel mapping

## Development Workflow

### Adding New Ports
1. Add port definition to `config.yaml` with correct `id`, `type`, `direction`
2. For DIN ports, specify `device` path; for USB ports, use ALSA client number as `id`
3. Port configs auto-created in main() via factory functions

### Adding Routes
1. Define inputs and outputs in config with optional channel numbers
2. Route objects created automatically from config in main()
3. Router handles all routing logic once started

### Testing MIDI Flow
1. Set `logLevel: "debug"` in config.yaml for message visibility
2. Connect USB/DIN devices and verify detection in logs
3. Send MIDI messages to input ports
4. Verify routing in logs: "Router received message" → "Routed message from X to Y"
5. Use `aconnect -l` to verify USB device client numbers

## Key Implementation Details

### Channel Processing
- MIDI channels stored as 1-16 in config, converted to 0-15 for message bytes
- Channel detection: `(msgByte & 0x0F) + 1` for messages 0x80-0xEF
- Channel setting: `(msgByte & 0xF0) | ((channel - 1) & 0x0F)`

### Goroutine Management
- Router starts one goroutine per input port for listening
- Additional goroutine per input to wrap raw messages with source info  
- Single `handleMessages()` goroutine processes all wrapped messages
- USB device detection runs in separate goroutine with event channel

### Memory Management
- Sender functions cached in router to avoid port reopening overhead
- Device repository updated in-place during USB polling
- Message channels use buffering to handle high-frequency MIDI data

## Instructions
- Use the custom logger instead of standard log package for consistency
- Always handle channels non-blocking with `select` to prevent deadlocks
- USB port IDs must be strings matching ALSA client numbers from `aconnect -l`
- When adding MIDI message processing, distinguish timing vs regular messages
- Follow the factory pattern for creating port configs (USB vs DIN)
- Test with real MIDI devices on Raspberry Pi for hardware-specific validation 
