
var baseSize = 130,
    minSize = 50,
    resizeRate = 1.5,
    throttledUpdateTextSize,
    submitting = false;

var msgNode, msgInNode;

function load() {
    msgNode = $("#msg");
    msgInNode = $("#msg-in");

    msgInNode.width($("#wrapper").width());
    msgInNode.keyup(function(e) {
        if (e.keyCode == 13 && !submitting) {
            submit();
        } else {
            updateText();
        }
    });

    msgInNode.click(function(e) {
        this.select();
    });

    throttledUpdateTextSize = _.throttle(updateTextSize, 1000);
    updateText();
}

function submit() {
    submitting = true;
    console.log("Submitting data");

    $.post("/submit", {msg: msgNode.text()})
        .success(redirect)
        .fail(function() {
            alert("Submission failed. Please try again.");
            submitting = false;
        });
}

function redirect(data) {
    if (data) {
        window.location = data;
    } else {
        alert("submission failed");
    }
}

function updateText() {
    var msg = msgInNode.val();
    msgNode.text(msg);
    throttledUpdateTextSize(msg);
}

function updateTextSize(msg) {
    msgNode.css("font-size", function() {
        var size = getFontSize(msg);
        return size;
    });
}

function getFontSize(msg) {
    var count = _.filter(msg, function(c) {
        return c != " ";
    }).length;
    return computeFontSize(count);
}

function computeFontSize(charCount) {
    var size =  Number.toInteger(baseSize - (charCount*resizeRate));
    return size > minSize ? size : minSize;
}


