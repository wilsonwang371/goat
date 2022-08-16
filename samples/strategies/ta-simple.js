addEventListener("onBars", function (bar) {
  res = talib.HtSine([
    1, 2, 3, 4, 5, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4,
    1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1,
    2, 3, 4,
  ]);
  console.log(res[0] + "  " + res[1]);
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

start_live()
