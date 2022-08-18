# GoAT
![Coverage](https://img.shields.io/badge/Coverage-46.3%25-yellow)

[![Go](https://github.com/wilsonwang371/goat/actions/workflows/go.yml/badge.svg)](https://github.com/wilsonwang371/goat/actions/workflows/go.yml)

This repo is currently a work in progress.

## Introduction

GoAT(Go Algo Trade) is inspired by PyAlgoTrade. It added support for live strategy execution. Currently it is under development.

## Design

There are several reasons for me proposing this GoAT as a GO alternative of PyAlgoTrade.

* Python code debugging at runtime is a headache
* Python is slow
* Dynamic typing is hard to debug
* PyAlgoTrade has not been updated for a long time.

However, I love using PyAlgoTrade. PyAlgoTrade is lightweight compared with Zipline. Zipline is not very flexible when
I want to make some small changes to meet my own needs.



## Build

```bash
run-build.sh build
```

## Test

```bash
run-build.sh test
```

## Run Strategy

### Live Mode

In live mode, the strategy will be executed in real time.

```bash

# This part is not complete yet.
./goat live -p fake -f samples/strategies/simple.js -S GLD

```

### Backtest Mode

In backtest mode, the strategy will be executed with historical data.

```bash

# Run a simple strategy with a csv file
./goat run -f samples/strategies/simple.js -s file://$(pwd)/samples/data/DBC-2007-yahoofinance.csv

# read data from yahoo finance
./goat run -f samples/strategies/simple.js -s remote://yahoo -S GLD

# By default, if the source is not an url, it will try to treat it at a file path.
./goat run -f samples/strategies/simple.js -s samples/data/DBC-2007-yahoofinance.csv

```


## Contributions

If you are interested in this project and want to contribute, please contact me.
