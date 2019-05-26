# Slackbot
[![GoDoc](https://godoc.org/github.com/c4milo/slackbot?status.svg)](https://godoc.org/github.com/c4milo/slackbot)

Primitive configuration management tool

## Usage

```
Usage:
  slackbot run <slackbook_file>
  slackbot -h | --help
  slackbot -v | --version

Commands:
  run                   Applies the state declared in the given slackbook YAML file

Options:
  -V --verbose          Turns on verbose output
  -h --help             Displays this help
  -v --version          Displays version string
```

## Example output
```
root@ubuntu:/mnt/hgfs/slackbot# ./bin/linux/slackbot run example/slackbook.yml
-----> Install apache2 and php5 packages...
-----> Make mod_dir try to find index.php first...
-----> Render index.php...
-----> Start apache2...
-----> Restart apache2...
-----> Reload apache2...
-----> Running 0 notified tasks...
Done 🎉
```
