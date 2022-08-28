var c = 0;
var lastTs = 0;

addEventListener("onBars", function (bar) {
  var thisTs = system.now();
  symbol = Object.keys(bar)[0];
  c++;

  ds = feed.dataseries(symbol, frequency.DAY, 10);

  if (thisTs - lastTs > 10) {
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

system.start();
