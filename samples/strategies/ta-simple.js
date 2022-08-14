addEventListener("onBars", function (bar) {
  console.log(talib.Atr(1, 2, 3));
  console.log(talib.Wma(1, 2, 3));
});

addEventListener("onStart", function () {
  // console.log("onStart is called.");
});

addEventListener("onFinish", function () {
  // console.log("onFinish is called.");
});

addEventListener("onIdle", function () {
  // console.log("onIdle is called.");
});
