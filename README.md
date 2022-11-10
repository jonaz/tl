# tl
time log - keep track on time worked and save state locally

```
$ tl --help
NAME:
   tl - A new cli application

USAGE:
   tl [global options] command [command options] [arguments...]

COMMANDS:
   calculate, c    Calculate worked time from start lunch-duration end
   in              stämpla in
   out             stämpla ut
   status, st      Print status of today or selected time range
   synclocklog, s  initiate interactive sync from locklog to tl json database
   undo            Undo last stamp
   help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --date value           Date (default: "2022-11-10")
   --file value           Logfile to save JSON data (default: "/var/log/time.log")
   --lock-log-file value  Logfile of computer lock history (default: "/var/log/i3lock")
   --help, -h             show help (default: false)
```
