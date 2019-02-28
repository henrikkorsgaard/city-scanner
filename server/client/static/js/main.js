document.addEventListener("DOMContentLoaded", bootstrap);

function bootstrap(){
	var running = document.querySelector("#wrapper .list#running")
	var archived = document.querySelector("#wrapper .list#archived")

	running.addEventListener("click", clicked)
	archived.addEventListener("click", clicked)


}


function clicked(e){
	console.log(e)
	var t = e.target.tagName === "H2" ? e.target.parentNode : e.target
	console.log(t)
	if(t.id === "running" || t.id === "archived" ){
		
		console.log(t)
		t.classList.toggle("open")
	}  else {

	}

}