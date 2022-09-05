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
  return [];
}

function getCloseSMA(ds, period) {
  var dsClosePrice = [];
  for (var i = 0; i < ds.length; i++) {
    dsClosePrice.push(ds[i].close);
  }
  if (dsClosePrice.length > period) {
    return talib.Sma(dsClosePrice, period);
  }
  return [];
}

var sma10, sma20, sma30, sma50, atr14, atr20;
var latestSma10,
  latestSma20,
  latestSma30,
  latestSma50,
  latestAtr14,
  latestAtr20;

addEventListener("onBars", function (bars) {
  var bar = bars[0];
  var thisTs = system.now();
  var symbol = Object.keys(bar);
  c++;

  var ds = getDataSeries(symbol, frequency.DAY, 64);
  if (ds == null) {
    return;
  }

  sma10 = getCloseSMA(ds, 10);
  sma20 = getCloseSMA(ds, 20);
  sma30 = getCloseSMA(ds, 30);
  sma50 = getCloseSMA(ds, 50);
  atr14 = getATR(ds, 14);
  atr20 = getATR(ds, 20);

  if (
    sma10 == null ||
    sma20 == null ||
    sma30 == null ||
    sma50 == null ||
    atr14 == null ||
    atr20 == null
  ) {
    return;
  }

  latestSma10 = sma10[sma10.length - 1];
  latestSma20 = sma20[sma20.length - 1];
  latestSma30 = sma30[sma30.length - 1];
  latestSma50 = sma50[sma50.length - 1];
  latestAtr14 = atr14[atr14.length - 1];
  latestAtr20 = atr20[atr20.length - 1];

  if (thisTs - lastTs > 10) {
    console.log("time: " + bar[symbol].dateTime);
    console.log("latestSma10: " + latestSma10);
    console.log("latestSma20: " + latestSma20);
    console.log("latestSma30: " + latestSma30);
    console.log("latestSma50: " + latestSma50);
    console.log("latestAtr14: " + latestAtr14);
    console.log("latestAtr20: " + latestAtr20);
    console.log(
      "[" + thisTs + "] onBars is called " + c + " times. Data: " + bar
    );
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

setInterval(function () {
  var res =
    "time: " + system.strftime("2006-01-02 15:04:05", system.now() + "\n");
  res += "sma10: " + latestSma10 + "\n";
  res += "sma20: " + latestSma20 + "\n";
  res += "sma30: " + latestSma30 + "\n";
  res += "sma50: " + latestSma50 + "\n";
  res += "atr14: " + latestAtr14 + "\n";
  res += "atr20: " + latestAtr20 + "\n";
  alert.info(res);
}, 1000 * 60 * 60 * 4);

system.start();
