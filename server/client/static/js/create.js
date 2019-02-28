document.addEventListener("DOMContentLoaded", bootstrap);

function bootstrap(){
	var search = document.querySelector("button#search")
	var geoloc = document.querySelector("button#geoloc")
	var create = document.querySelector("button#create")
	var cancel = document.querySelector("button#cancel")

	var coords = {
		lat: 0,
		lng: 0,
	}
	
	if (!navigator.geolocation){
		geoloc.style.display = "none"
	}

	var map = L.map("mapview").setView([51.505, -0.09], 4);
	var tiles = L.tileLayer('https://tiles.wmflabs.org/bw-mapnik/{z}/{x}/{y}.png', 
		{
				maxZoom: 25,
				attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
		});
	map.addLayer(tiles)

	map.addEventListener("moveend", function(e){
		coords.lat = map.getCenter().lat
		coords.lng = map.getCenter().lng		
	})

	map.addEventListener("zoomend", function(e){
		coords.lat = map.getCenter().lat
		coords.lng = map.getCenter().lng	
	})

	cancel.addEventListener('click', function(e){
		e.preventDefault()
		document.location.replace("/")
	});

	geoloc.addEventListener('click', function(e){
		e.preventDefault()
		if (navigator.geolocation) {
			navigator.geolocation.getCurrentPosition(function(pos){
				
				coords.lat = pos.coords.latitude
				coords.lng = pos.coords.longitude
				map.setView([pos.coords.latitude, pos.coords.longitude], 12)
				
			});
		}
	});

	search.addEventListener('click', function(e){
		e.preventDefault()
		var data = document.querySelector("#city")
		if(data.value.trim() !== ""){
			apiRequest(data.value, function(data){
				if(data && data.results.length > 0){
					
					var results = data.results
					var found = results[0]
					for(var i = 0, n = results.length; i < n; i++){
						var res = results[i];
						if(res["_type"] === "city"){
							found = res;
							break;
						}
					}
					coords.lat = found.geometry.lat
					coords.lng = found.geometry.lng
					map.setView([found.geometry.lat, found.geometry.lng], 12)

				} else {
					console.log("error")
				}
			})
		}
		console.log(data.value)
	});

	create.addEventListener('click', function(e){
		e.preventDefault()
		//validate
		//post data to server
		//redirect to experiment template with generated files for nodes.

	})
}

function apiRequest(location, callback){
	var xhr = new XMLHttpRequest();
	//This need to happen serverside bro
	xhr.open('GET', 'https://api.opencagedata.com/geocode/v1/json?q='+location+'&key=a4008935b3ef46faa7a43143a98bd37e&language=en&pretty=1&no_annotations=1');
	xhr.onreadystatechange = function () {
		if (xhr.readyState === 4 && xhr.status === 200) {
			try {
				var result = JSON.parse(xhr.responseText)
				callback(result)
			} catch(e){
				callback()
			}
		}
	}
	xhr.send(null);
}
