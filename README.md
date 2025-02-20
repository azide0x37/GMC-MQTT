# GMC-MQTT

GMC-MQTT is a simple Golang application designed for Raspbian that connects to a GMC-300s device via USB-serial, queries it for data, and publishes that data to an MQTT topic. The application uses a TOML configuration file to set options such as MQTT host, port, query interval, and serial port settings. It also subscribes to MQTT topics for both temporary and permanent configuration updates.

## Features

- **Serial Communication:** Connects to a GMC-300s over USB-serial.
- **Periodic Querying:** Interrogates the device at a configurable interval.
- **MQTT Integration:** Publishes device data and listens for config update messages.
- **TOML Configuration:** Easily change options via a simple config file.
- **Modular and Testable:** Code segmented into clear packages with a sample test.

## Installation

For a simple deployment, run the following command on your Raspbian device:

```sh
curl -sL https://derahm.com/gmc-mqtt-install.sh | sh
```

## Build Procedure
1. Clone the repository:

    ```sh
    git clone https://github.com/azide0x37/GMC-MQTT.git
    cd gmc-mqtt
    ```

2. Download dependencies and build:
    ```sh
    go mod tidy
    go build -o gmc-mqtt ./cmd
    ```

1.  Run tests:

    ```sh
    go test ./...
    ```

## Configuration
Edit the config.toml file to suit your setup:

```toml
serial_device = "/dev/ttyUSB0"
baud_rate = 9600
mqtt_host = "localhost"
mqtt_port = 1883
query_interval = 10
publish_topic = "gmc/data"
config_topic = "gmc/config/temp"
permanent_config_topic = "gmc/config/permanent"
```

## Running
After building the application, run it as follows:

```sh
./gmc-mqtt -config config.toml
```

## Future Improvements
Enhance error handling and automatic reconnection.
Implement persistent configuration updates.
Support more advanced command parsing for device interrogation.


## License
MIT License