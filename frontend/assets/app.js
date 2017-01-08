function toggleVisibility(target) {
  target.parentElement.classList.toggle('hide-details');
}

function timeSince(timestamp) {
  var date = new Date(timestamp * 1000);
  var seconds = Math.floor((new Date() - date) / 1000);
  var interval = Math.floor(seconds / 31536000);
  if (interval > 1) {
      return interval + " years";
  }
  interval = Math.floor(seconds / 2592000);
  if (interval > 1) {
      return interval + " months";
  }
  interval = Math.floor(seconds / 86400);
  if (interval > 1) {
      return interval + " days";
  }
  interval = Math.floor(seconds / 3600);
  if (interval > 1) {
      return interval + " hours";
  }
  interval = Math.floor(seconds / 60);
  if (interval > 1) {
      return interval + " minutes";
  }
  return Math.floor(seconds) + " seconds";
}

function contextClassName(lineNoA, lineNoB) {
  return (lineNoA == lineNoB) ? "highlight" : "nohighlight";
}

function deleteError(errorNo) {
  var xhr = new XMLHttpRequest();
  xhr.onreadystatechange = function() {
    if (xhr.readyState === 4) {
      var element = document.getElementById("error-" + errorNo);
      element.parentElement.removeChild(element);
    }
  }
  xhr.open('DELETE', '/errors/' + errorNo);
  xhr.send();
}

(function(d) {
  var xhr = new XMLHttpRequest();
  xhr.onreadystatechange = function() {
    if (xhr.readyState === 4) {
      var response = JSON.parse(xhr.responseText);
      d.getElementById("api-key").innerHTML = response.apiKey;
      d.getElementById("main").innerHTML = tmpl("errors", response.errors);
    }
  };
  xhr.open('GET', '/errors');
  xhr.send();
})(document);
