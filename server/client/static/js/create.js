document.addEventListener("DOMContentLoaded", bootstrap);

function bootstrap(){
	var search = document.querySelector("button#search")
	var geoloc = document.querySelector("button#geoloc")
	var create = document.querySelector("button#create")
	var cancel = document.querySelector("button#cancel")
	var name_err = document.querySelector("p.name-err")
	var email_err = document.querySelector("p.email-err")

	var name = document.querySelector("input#name")
	var email = document.querySelector("input#email")

	var emailRegEx = new RegExp("[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*@(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?")

	var coords = {
		lat: 4,
		lng: 4,
		zoom: 12
	}

	var validation = {
		email:false,
		name:false,
		coordinates:false
	}

	validate();

	email.addEventListener('blur', validateEmail)
	name.addEventListener('blur', validateName)

	function validateEmail(){
		if(emailRegEx.test(email.value)){
			validation.email = true;
			email.classList.remove("error")
			email_err.style.visibility = "hidden"
		} else {
			email_err.style.visibility = "visible"
			email_err.innerHTML = "Must provide email"
			email.classList.add("error")
			validation.email = false;
		}
		validate()
	}

	validateEmail()

	function validateName(){
		if(name.value.trim() !== ""){
			isNameUnique(name.value, function(result){
				if(result){
					name.classList.remove("error")
					validation.name = true;
					name_err.style.visibility = "hidden"
				} else {
					name_err.style.visibility = "visible"
					name_err.innerHTML = "Name already taken"
					name.classList.add("error")
					validation.name = false;
				}
				validate()
			})
		
		} else {
			name_err.style.visibility = "visible"
			name.classList.add("error")
			name_err.innerHTML = "Must provide name"
			validation.name = false;
			validate()
		}
		
	}

	validateName()

	function validate(){
		console.log(validation)
		if(validation.name && validation.email && validation.coordinates){
			create.disabled = false;
		} else {
			create.disabled = true;
		}
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
		coords.zoom = map.getZoom()	
		validation.coordinates = true;
		validate();
	})

	map.addEventListener("zoomend", function(e){
		coords.lat = map.getCenter().lat
		coords.lng = map.getCenter().lng
		coords.zoom = map.getZoom()
		validation.coordinates = true;
		validate();
	})

	cancel.addEventListener('click', function(e){
		e.preventDefault()
		document.location.replace("/")
	});

	geoloc.addEventListener('click', function(e){
		document.querySelector(".lds-ellipsis").classList.add("active")
		e.preventDefault()
		if (navigator.geolocation) {
			navigator.geolocation.getCurrentPosition(function(pos){
				
				coords.lat = pos.coords.latitude
				coords.lng = pos.coords.longitude
				coords.zoom = 12
				map.setView([pos.coords.latitude, pos.coords.longitude], 12)
				document.querySelector(".lds-ellipsis").classList.remove("active")
				validation.coordinates = true;
				validate();
			});
		}
	});

	search.addEventListener('click', function(e){
		document.querySelector(".lds-ellipsis").classList.add("active")
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
					coords.zoom = 12
					map.setView([found.geometry.lat, found.geometry.lng], 12)
					document.querySelector(".lds-ellipsis").classList.remove("active")
					validation.coordinates = true;
					
					validate();

				} else {
					console.log("error")
				}
			})
		}
		console.log(data.value)
	});



	create.addEventListener('click', function(e){
		e.preventDefault();
		console.log(coords.lat)
		postData({email: email.value, name: name.value, lat: coords.lat, lng: coords.lng, zoom: coords.zoom})
	})
}

function postData(data){
	var xhr = new XMLHttpRequest();
	
	xhr.open('POST', './create');
	
	xhr.onreadystatechange = function () {
		if (xhr.readyState === 4 && xhr.status === 200) {
			document.location = xhr.responseText
		} else {
			console.log("HENRIK NEED TO FIX THIS")
		}
	}
	console.log(JSON.stringify(data))
	xhr.send(JSON.stringify(data));
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

function isNameUnique(name, callback){
	var xhr = new XMLHttpRequest();
	//This need to happen serverside bro
	xhr.open('GET', './validate/name/'+name);
	xhr.onreadystatechange = function () {
		if (xhr.readyState === 4 && xhr.status === 200) {
			callback(false)
		} else if (xhr.readyState === 4 && xhr.status === 404){
			callback(true)
		}
	}
	xhr.send(null);
}
