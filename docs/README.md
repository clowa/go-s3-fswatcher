# Overview

Developers notes for the S3 File Watcher project.

## Events fired by certain filesystem operations

Cmd: ```touch test.txt```
Events:

```txt
WRITE         "watch/test.txt"`
```

Cmd: ```echo "Hello World" > test.txt````
Event:

```txt
Event: WRITE         "watch/test.txt"
Event: WRITE         "watch/test.txt"
Event: WRITE         "watch/test.txt"
```

Cmd: ```mv test.txt demo.txt```
Event:

```txt
2025/04/06 12:00:08 Starting S3 File Watcher
Event: RENAME        "watch/test.txt"
Event: WRITE         "watch/demo.txt"
```

Cmd: ```nano test.txt```
Event:

```txt
Event: WRITE         "watch/.test.txt.swp"
Event: WRITE         "watch/.test.txt.swp"  - periodicly, maybe auto-save
Event: WRITE         "watch/test.txt"       - on Ctrl+O
```
