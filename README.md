# event-messenger

dictyBase server to handle events as a subscriber through Nats messaging.

# Available commands

```
NAME:
   event-messenger - Handle events from nats messaging

USAGE:
   event-messenger [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
     gh-issue    creates a github issue when a new stock order comes through
     send-email  sends an email when a new stock order comes through
     help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --log-format value  format of the logging out, either of json or text. (default: "json")
   --log-level value   log level for the application (default: "error")
   --help, -h          show help
   --version, -v       print the version
```
