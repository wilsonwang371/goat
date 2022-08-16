var c = 0;

addEventListener("onBars", function (bar) {
  c++;
  console.log("onBars is called " + c + " times. Data: " + bar);
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

system.start()
