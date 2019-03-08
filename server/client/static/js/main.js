document.addEventListener("DOMContentLoaded", bootstrap);

function bootstrap(){
	var active = document.querySelector("#wrapper .list#active")
	var archived = document.querySelector("#wrapper .list#archived")

	active.addEventListener("click", clicked)
	archived.addEventListener("click", clicked)


}


function clicked(e){
	console.log(e)
	var t = e.target.tagName === "H2" ? e.target.parentNode : e.target
	console.log(t)
	if(t.id === "active" || t.id === "archived" ){
		
		console.log(t)
		t.classList.toggle("open")
	}  else {

	}

}