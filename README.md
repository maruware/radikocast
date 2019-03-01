# radikocast

Record radiko programs, and publish your private podcast.

## Usage

```bash
$ radikocast help                                                                                                                                                                                                               
Usage: radikocast [--version] [--help] <command> [<args>]

Available commands are:
    publish     Publish podcast
    rec         Record a radiko program
    rss         Generate podcast RSS
    schedule    Schedule programs
```

All commands need config yaml file.
### config yaml format

See [config.sample.yml](config.sample.yml)

Currently, only s3 is supported for publishing.  
The folloing environment variables are required for publishing to s3.

* AWS_ACCESS_KEY_ID
* AWS_SECRET_ACCESS_KEY
* AWS_REGION

or setting IAM Role.

### schedule

Schedule recording and publish reservations.

```bash
$ radikocast schedule -h
Usage: radikocast schedule [options]
  Schedule programs
Options:
  -config,c=filepath       Config file path (default: config.yml)
```

### rec

Record a radiko program.

```bash
$ radikocast rec -h
Usage: radikocast rec [options]
  Record a radiko program.
Options:
  -id=name                 Station id
  -start,s=201610101000    Start time
  -area,a=name             Area id
  -config,c=filepath       Config file path (default: config.yml)
```

### rss

Generate podcast RSS file.

```bash
$ radikocast rss -h
Usage: radikocast rss [options]
  Generate podcast RSS
Options:
  -config,c=filepath       Config file path (default: config.yml) 
```

### publish

```bash
$ radikocast publish -h
Usage: radikocast publish
  Publish podcast
Options:
  -config,c=filepath       Config file path (default: config.yml)
```

## Installation

Download binary from [release page](https://github.com/maruware/radikocast/releases) and place it in $PATH directory.

### Requirements

* ffmpeg