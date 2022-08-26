var c = 0;

addEventListener("onBars", function (bar) {
  symbol = Object.keys(bar)[0];
  c++;
  console.log("onBars is called " + c + " times. Data: " + bar);
  console.log("dataseries: " + feed.dataseries(symbol, 86400, 10));
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
