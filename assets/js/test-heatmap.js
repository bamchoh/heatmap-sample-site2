function load_clickpoint(img) {
  fetch('/api/click' + window.location.search)
    .then(function(response){
      return response.json();
    })
    .then(function(my_json) {
      if(my_json == null || my_json["points"] == null) {
        const loaded = document.getElementById("loading");
        loaded.classList.add("loaded");
        const errorDiv = document.getElementById("error");
        errorDiv.innerText = "クリックポイントはまだありません";
        errorDiv.classList.remove("hidden");
        return;
      }

      var heatmap_div = document.querySelector('.heatmap');
      var heatmapInstance = h337.create({
        container: heatmap_div,
        maxOpacity: .9,
      });

      var points = [];
      my_json["points"].forEach(point => {
        var point = {
          x: Math.floor(point.x * img.clientWidth),
          y: Math.floor(point.y * img.clientHeight),
          value: 1,
          radius: 35
        };
        points.push(point);
      });
      heatmapInstance.addData(points);

      const loaded = document.getElementById("loading");
      loaded.classList.add("loaded");
    });
}

function pngpath() {
  var host = '/api/png/'
  if(window.location.search) {
    return host + window.location.search + '&width=' + window.innerWidth
  } else {
    return host
  }
}

fetch(pngpath())
  .then(function(response) {
    return response.json();
  })
  .then(function(my_json) {
    if(my_json == null) {
      const loaded = document.getElementById("loading");
      loaded.classList.add("loaded");
      return
    }

    if(my_json["status"] == "error") {
      const loaded = document.getElementById("loading");
      loaded.classList.add("loaded");
      return
    }

    if(my_json["png"] == null) {
      const loaded = document.getElementById("loading");
      loaded.classList.add("loaded");
      return
    }

    var img = new Image();
    img.src = 'data:image/png;base64,' + my_json["png"]

    var div = document.querySelector('.heatmap');
    div.appendChild(img);

    load_clickpoint(img);
  })

