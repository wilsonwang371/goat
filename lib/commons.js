function fetchDataSeries(sym, freq, len) {
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

function calcATR(ds, period) {
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

function calcCloseSMA(ds, period) {
  var dsClosePrice = [];
  for (var i = 0; i < ds.length; i++) {
    dsClosePrice.push(ds[i].close);
  }
  if (dsClosePrice.length > period) {
    return talib.Sma(dsClosePrice, period);
  }
  return [];
}

module.exports = {
  fetchDataSeries: fetchDataSeries,
  calcATR: calcATR,
  calcCloseSMA: calcCloseSMA,
};
