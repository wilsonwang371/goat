addEventListener("onBars", function (bar) {
  console.log("" + talib.Wma([1, 2, 3, 4, 5], 4));
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
