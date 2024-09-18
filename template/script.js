function showSuggestions(query) {
    if (query.length == 0) {
        document.getElementById("suggestions").innerHTML = "";
        return;
    }
    
    fetch(`/search-query?s=${query}`)
        .then(response => response.json())
        .then(data => {
            let suggestions = data ? data.map(item => 
                `<div><a href="/profil?id=${item.id}">${item.name} - ${item.type}</a></div>`
            ).join('') : "not found";
            document.getElementById("suggestions").innerHTML = suggestions;
        });
}
