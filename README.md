# GoAlgoTrade

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
./build.sh
```

## Run

```bash
./goalgotrade
```

### Run Strategy JS Script

```bash
./goalgotrade run -s sample/simple.js
```


## Contributions

If you are interested in this project, please contact me.
