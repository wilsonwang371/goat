
addEventListener("onBars", function(bar){
    console.log("onBars is called. " + bar);
});

addEventListener("onStart", function(){
    console.log("onStart is called.");
});

addEventListener("onFinish", function(){
    console.log("onFinish is called.");
});

addEventListener("onIdle", function(){
    console.log("onIdle is called.");
});
