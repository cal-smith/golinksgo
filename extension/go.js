browser.webRequest.onBeforeRequest.addListener((request) => {   
    const url = request.url.replace("%2F", "/");
    let [query] = url.match(/go\/.+/i) ?? [];
    if (!query) {
        [query] = url.match(/go\/.+/i) ?? []
    }
    if (!query && (/go\//i).test(url)) {
        console.log(request.url.replace("%2F", "/"));
        return {redirectUrl: `http://go.to/`};
    }
    if (query) {
        console.log(request.url.replace("%2F", "/"));
        const rawLink = query.split("&")[0];
        const path = rawLink.split("/")[1];
        return {redirectUrl: `http://go.to/${path}`};
    }
}, {urls: ["<all_urls>"]}, ["blocking"]);

const linkToSuggestion = (link) => {
    const source = link.Source.substring(1);
    return {
        content: source,
        description: `go/${source}`,
    };
};

browser.omnibox.onInputChanged.addListener((search, suggest) => {
    search = search.trim();
    
    if (search === "") {
        return;
    }

    fetch(`http://go.to/api/search?q=${search}`)
        .then(res => res.json())
        .then(data => suggest(data.map(linkToSuggestion)));
});

browser.omnibox.onInputEntered.addListener((text, disposition) => {
    const url = `http://go.to/${text}`;
    switch (disposition) {
        case "currentTab":
          browser.tabs.update({url});
          break;
        case "newForegroundTab":
          browser.tabs.create({url});
          break;
        case "newBackgroundTab":
          browser.tabs.create({url, active: false});
          break;
      }
});
