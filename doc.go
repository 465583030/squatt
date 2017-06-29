//go:generate .script/doc.sh

// The SQuaTT MQTT Server
//
// Usage:
//   squatt [flags]
//
// Flags:
//       --config string         Config file (default "$HOME/.squatt.yml")
//       --data string           Data folder (default "$HOME/.squatt")
//       --debug                 Use debug mode
//   -h, --help                  help for squatt
//       --listen string         Port to listen on (default ":1883")
//       --listen-debug string   Debug port to listen on (if debug mode on) (default ":6060")
package main
