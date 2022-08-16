# GoAlgoTrade

[![Go](https://github.com/wilsonwang371/goalgotrade/actions/workflows/go.yml/badge.svg)](https://github.com/wilsonwang371/goalgotrade/actions/workflows/go.yml)

This repo is currently a work in progress.

## Introduction

GoAlgoTrade is a Go implementation of PyAlgoTrade. Currently it is under development.

## Design

There are several reasons for me proposing this GoAlgoTrade as a GO alternative of PyAlgoTrade.

* Python code debugging at runtime is a headache
* Python is slow
* Dynamic typing is hard to debug
* PyAlgoTrade has not been updated for a long time.

However, I love using PyAlgoTrade. PyAlgoTrade is lightweight compared with Zipline. Zipline is not very flexible when
I want to make some small changes to meet my own needs.



## Build

```bash
build.sh build
```

## Test

```bash
build.sh test
```

## Run Strategy


### Run A Simple Strategy

```bash

# Run a simple strategy with a csv file
./goalgotrade run -f samples/strategies/simple.js -s file://$(pwd)/samples/data/DBC-2007-yahoofinance.csv

# read data from yahoo finance
./goalgotrade run -f samples/strategies/simple.js -s remote://yahoo -S GLD

# By default, if the source is not an url, it will try to treat it at a file path.
 ./goalgotrade run -f samples/strategies/simple.js -s samples/data/DBC-2007-yahoofinance.csv

```

### Run A Strategy Live

```bash
./goalgotrade live -p fake -f samples/strategies/simple.js -S XAUUSD
```


## Contributions

If you are interested in this project and want to contribute, please contact me.
