var c = 0;
var lastTs = 0;

function getDataSeries(sym, freq, len) {
  var ds = feed.dataseries(sym, freq, len);
  if (ds == null) {
    console.log("No data series for " + sym + " at frequency " + freq);
    return;
  }
  if (Object.keys(ds).length == 0) {
    console.log("ds is null");
    return;
  }
  return ds.data;
}

function getATR(ds, period) {
  var dsHighPrice = [];
  for (var i = 0; i < ds.length; i++) {
    dsHighPrice.push(ds[i].high);
  }
  var dsLowPrice = [];
  for (var i = 0; i < ds.length; i++) {
    dsLowPrice.push(ds[i].low);
  }
  var dsClosePrice = [];
  for (var i = 0; i < ds.length; i++) {
    dsClosePrice.push(ds[i].close);
  }
  if (dsClosePrice.length > period) {
    return talib.Atr(dsHighPrice, dsLowPrice, dsClosePrice, period);
  }
  return null;
}

function getCloseSMA(ds, period) {
  var dsClosePrice = [];
  for (var i = 0; i < ds.length; i++) {
    dsClosePrice.push(ds[i].close);
  }
  if (dsClosePrice.length > period) {
    return talib.Sma(dsClosePrice, period);
  }
  return null;
}

addEventListener("onBars", function (bars) {
  var bar = bars[0];
  var thisTs = system.now();
  var symbol = Object.keys(bar);
  c++;

  var ds = getDataSeries(symbol, frequency.DAY, 64);
  if (ds == null) {
    return;
  }

  var sma10 = getCloseSMA(ds, 10);
  var sma20 = getCloseSMA(ds, 20);
  var sma30 = getCloseSMA(ds, 30);
  var sma50 = getCloseSMA(ds, 50);
  var atr14 = getATR(ds, 14);
  var atr20 = getATR(ds, 20);

  if (thisTs - lastTs > 10) {
    console.log("SMA(10): " + sma10);
    console.log("SMA(20): " + sma20);
    console.log("SMA(30): " + sma30);
    console.log("SMA(50): " + sma50);
    console.log("ATR(14): " + atr14);
    console.log("ATR(20): " + atr20);
    console.log(
      "[" + thisTs + "] onBars is called " + c + " times. Data: " + bar
    );
    // console.log("Data series: " + JSON.stringify(ds));
    lastTs = thisTs;
  }
});

addEventListener("onStart", function () {
  console.log("onStart is called.");
});

addEventListener("onFinish", function () {
  console.log("onFinish is called.");
});

addEventListener("onIdle", function () {
  // console.log("onIdle is called.");
});

system.start();
