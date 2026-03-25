package mqtt

import paho "github.com/eclipse/paho.mqtt.golang"

// Re-export Paho types for use by callers of this package.
type Client = paho.Client

type MessageHandler = paho.MessageHandler
