
addEventListener("onBars", function (bar) {
  var value = kvstorage.load("counter");
  if (value == null) {
    console.log("no previous value");
    value = 0;
  } else {
    value = parseInt(value);
  }
  value = value + 1;
  kvstorage.save("counter", ''+value);
  console.log("onBars is called " + value + " times. Data: " + bar);
});

addEventListener("onStart", function () {
  console.log("onStart is called.");
});

addEventListener("onFinish", function () {
  console.log("onFinish is called.");
});

addEventListener("onIdle", function () {
  //console.log("onIdle is called.");
});

start_live()
