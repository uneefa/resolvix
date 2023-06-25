# Resolvix

Small, fast, simple tool for checking a list of dns resolvers to see which ones are alive.

## Installation

```sh
go install github.com/uneefa/resolvix@latest
```

## Usage

```sh
root~$ resolvix -list resolvers.txt
4.2.2.2
1.1.1.1
8.8.8.8
216.146.36.36
8.8.4.4
1.0.0.1
4.2.2.1
216.146.35.35
208.67.220.220
208.67.222.222
...
```

## Parameters

```sh
Usage:
  resolvix -list resolvers.txt -output alive.txt
  # or
  cat resolvers.txt | resolvix -output alive.txt

Options:
  -list     string   List of DNS resolvers
  -output   string   Output file
  -protocol string   Network protocol (default "udp")
  -silent            Silent mode
  -timeout  int      Timeout in seconds (default 1)
  -workers  int      Number of workers (default 10)
  -h                 Show this help message
```
