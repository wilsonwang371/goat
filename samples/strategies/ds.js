var c = 0;

addEventListener("onBars", function (barStr) {
  bar = JSON.parse(barStr);
  symbol = Object.keys(bar)[0];
  c++;
  console.log("onBars is called " + c + " times. Data: " + bar);
  feed.dataseries(symbol, 0);
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
