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
    rec_schedule    Rec a program by schedule expression.
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
  -bucket=bucketname	   S3 bucket name
```

### rec_schedule

Record a radiko program by schedule expression.

```bash
$ radikocast rec_schedule -h
Usage: radikocast rec_schedule [options]
  Record a radiko program.
Options:
  -id=name                 Station id
  -day=wednesday           everyday or weekday, sunday, monday, ..., saturday
  -at=13:00
  -area,a=name             Area id
  -bucket=bucketname	   S3 bucket name
```

### rss

Generate podcast RSS file.

```bash
$ radikocast rss -h
Usage: radikocast rss [options]
  Generate podcast RSS
Options:
  -title title
  -host host
  -image image
  -bucket bucket
  -feed feed
```

## Installation

Download binary from [release page](https://github.com/maruware/radikocast/releases) and place it in $PATH directory.

### Requirements

* ffmpeg