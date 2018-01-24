# GPM - The lightweight process manager
GPM is an extremely lightweight process manager.  
It's easy to configure and has no external dependencies, making it an obvious choice for contains.  
GPM runs on any Platform, not just Linux or Windows. 

## Features
GPM comes with a pretty short list of core features:

### Automatic restart
Automatically restart a process whenever it dies.

### Inter-dependencies
Have a setup script that needs to run before other process can run? 
Just have them depend on each other, and GPM will handle the rest


### Graceful shutdown
GPM will pass along interrupts, and give child processes a chance to shutdown. Are they not down 
after 7 seconds, then they will be force killed. 

### Stdout/stderr handling
All output to stdout/stderr from child-processes will appear in GPM's stdout/stderr, allowing 
for better log tailing when running in contains, or simply for easier overview.

## Configuration
GPM attempts to keep the configuration to a minimum, however some configuration is required for 
GPM to be able to figure out what processes should be run. 

Make a file `config.json` where you want to run GPM from.

An extremely simply configuration file, that just runs an echo command once looks like this:
```json
[
  {
    "name": "echo",
    "command": "echo 'this is a test'"
  }
]
```

Here is a table of all the possible options per process.

|Key|Description|Required|
|------|-----|------|
|`name`|This is the name of the process, used when resolving dependencies, and for writing to the log.|Yes|
|`command`|This is the actual terminal command to run. Write here exactly like you would on your normal terminal. Does not support piping between processes.|Yes|
|`autoRestart`|Set to true to have the process automatically be restarted when it closes. Mutually exclusive with `after`|No|
|`after`|The name of the process, **this** process should be run after. Mutually exclusive with `autoRestart`|No|
|`workDir`|The working directory of the process when executed|No|

A more involved example:
```json
[
  {
    "name": "echo",
    "command": "echo 'this is a test'"
  },
  {
    "name": "gfs",
    "command": "gfs-windows-x64.exe",
    "autoRestart": true,
    "after": "echo"
  },
  {
    "name": "echo2",
    "command": "echo 'this is echo 2'",
    "after": "echo"
  },
  {
    "name": "echo3",
    "command": "echo this is echo 3",
    "after": "echo2"
  }
]
```

This starts a single echo process that write `'this is a test'` to the terminal. 
Then it starts a [GFS](https://github.com/zlepper/gfs) process. 
At the same time another echo process is started, writing `'this is echo 2'` to the terminal.
Then yet another echo process start, that write `'this is echo 3'` to the terminal.

Should the GFS process ever stop, then GPM will handle starting it again. 

