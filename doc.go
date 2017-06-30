//go:generate .script/doc.sh

// The SQuaTT MQTT Server
//
// Usage:
//   squatt [flags]
//
// Flags:
//       --config string            Config file (default "$HOME/.squatt.yml")
//       --data string              Data folder (default "$HOME/.squatt")
//       --debug                    Debug mode
//   -h, --help                     help for squatt
//       --listen.debug string      Debug server listen address (default "127.0.0.1:6060")
//       --listen.tcp string        MQTT server TCP listen address (default ":1883")
//       --listen.tls string        MQTT server TLS listen address
//       --tls.certificate string   Path to certificate for TLS (default "cert.pem")
//       --tls.key string           Path to private key for TLS (default "key.pem")
package main
