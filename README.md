# GoAT
[![Compile and Test](https://github.com/wilsonwang371/goat/actions/workflows/basic.yml/badge.svg)](https://github.com/wilsonwang371/goat/actions/workflows/basic.yml)
![Coverage](https://img.shields.io/badge/Coverage-50.6%25-yellow)

## Introduction

GoAT(Go Algo Trade) is inspired by PyAlgoTrade. It added support for live strategy execution. 

## Design

There are several reasons for me proposing this GoAT as a GO alternative of PyAlgoTrade.

* Python code debugging at runtime is a headache
* Python is slow
* Dynamic typing is hard to debug
* PyAlgoTrade has not been updated for a long time.

However, I love using PyAlgoTrade. PyAlgoTrade is lightweight compared with Zipline. Zipline is not very flexible when
I want to make some small changes to meet my own needs.

## Features

GoAT currently supports the following features:

* User strategy Javascript support
* Multiple data frequencies support
* Live strategy support
* Data format conversion support

Broker, strategy & portfolio analysis are not there yet. I will add them in the future.

## Build

```sh

# On Linux
run-build.sh compile

# On MacOS
make compile

```

## Test

```
run-build.sh test
```

## Format Code

```sh

# format go and js code
run-build.sh format
```

## Run Strategy

### Live Mode

In live mode, the strategy will be executed in real time.

```sh

# Generate some fake data
./goat live -p fake -f samples/strategies/simple.js -S GLD

# Run strategy with real data
./goat live -p goldpriceorg -f samples/strategies/simple.js -S XAUUSD

# Multi providers are also supported
./goat live -p "goldpriceorg,fake" -f samples/strategies/simple.js -S XAUUSD

# Run recovery mode
./goat live -p "goldpriceorg,fake" -f samples/strategies/simple.js -S XAUUSD \
    -r samples/data/strategy_data.dumpdb

```

### Backtest Mode

In backtest mode, the strategy will be executed with historical data.

```sh

# Run a simple strategy with a csv file
./goat run -f samples/strategies/simple.js -s \
    file://$(pwd)/samples/data/DBC-2007-yahoofinance.csv

# read data from yahoo finance
./goat run -f samples/strategies/simple.js -s remote://yahoo -S GLD

# By default, if the source is not an url, it will try to treat it at a file path.
./goat run -f samples/strategies/simple.js -s \
    samples/data/DBC-2007-yahoofinance.csv

```

### Convert Other DB to GoAT sqlite DB

```sh

# convert a sqlite db to GoAT sqlite db
./goat convert -f samples/convert/mappings.js -s ./samples/data/strategy_data.sqlite -t sqlite \
    -o ./samples/data/strategy_data.dumpdb

```


## Contributions

If you are interested in this project and want to contribute, please contact me.
