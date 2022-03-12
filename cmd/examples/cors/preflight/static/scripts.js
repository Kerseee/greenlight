document.addEventListener("DOMContentLoaded", function(){
    fetch("http://localhost:8080/v1/tokens/authentication", {
        method: "POST",
        headers: {
            'Content-Type':'application/json'
        },
        body: JSON.stringify({
            email: 'raichu@example.com',
            password: 'raichuraichu'
        })
    }).then(
        function(resp){
            resp.text().then(function(text){
                document.getElementById("output").innerHTML = text;
            });
        },
        function(err){
            document.getElementById("output").innerHTML = err;
        }
    );
});