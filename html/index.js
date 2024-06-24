const searchElement = document.getElementById('search');
const resultsElement = document.getElementById('results');
const searchContainer = document.getElementById('searchContainer');
const wildcardElement = document.getElementById('wildcard');
const body = document.getElementsByTagName('body')[0];

wildcardElement.addEventListener('change', () => {
    searchElement.dispatchEvent(new Event('input'));
})

searchResults = []
showPermissions = true
headItem = 0
const maxItems = 200
const minSearchLength = 4

const stringToBoolean = (stringValue) => {
    switch(stringValue?.toLowerCase()?.trim()){
        case "true": 
        case "yes": 
        case "1": 
          return true;

        case "false": 
        case "no": 
        case "0": 
        case null: 
        case undefined:
          return false;

        default: 
          return JSON.parse(stringValue);
    }
}

    const searchData = (event) => {
        const value = event.target.value;
        headItem = 0
        if (value.length < minSearchLength) {
            if (value.length == 0) {
                resultsElement.innerHTML = ""
                body.className = "bodyWithoutResults"
            } else {
                resultsElement.innerHTML = `<div>Please type at least ${minSearchLength} characters of the permission name</div>`
            }
            return
        }
        let res = fetch(`/query?qp=${value}&wildcard=${wildcardElement.checked}`)
            .then((res) => {
                jsonPromise = res.json()
                jsonPromise.then((data) => {
                    console.log(`fetch complete: result size: ${data.length}`)
                    searchResults = data
                    if (data.length == 0) {
                        resultsElement.innerHTML = "<div>No data</div>"
                        body.className = "bodyWithoutResults"
                        console.log("No data")
                        return
                    }
                    showPermissions = wildcardElement.checked
                    permissionsClass = showPermissions ? "permCell" : "permCellHidden"

                    data = searchResults.slice(headItem, headItem + maxItems)
                    body.className = "bodyWithResults"
                    resultCountDiv = `<div>(Showing ${maxItems} of ${searchResults.length} results)</div>`
                    if (searchResults.length <= maxItems) {
                        resultCountDiv = ""
                    }
                    header = `<div class="resultsTable">
                                <div class="resultsRow resultsHeader">
                                    <div class="${permissionsClass}">Permission</div>
                                    <div class="roleCell">Role</div>
                                </div> 
                        `;
                    footer = `</div>`
                    rows = data.map((item) => {
                        role = item.role.replace(/^(roles\/)/,"");
                        return `
                        <div class="resultsRow">
                            <div class="${permissionsClass}">${item.permission}</div>
                            <div class="roleCell">
                                <a href="https://cloud.google.com/iam/docs/understanding-roles#${role}" 
                                    target="_blank"
                                    title="View documenation for role ${role}"
                                    aria-label="View documenation for role ${role}"
                                    >
                                    ${role}
                                </a>
                            </div>
                        </div> `;
                    }).join('');
                    resultsElement.innerHTML = header + rows + resultCountDiv + footer
                });
            })
}

const debounce = (callback, waitTime) => {
    let timer;
    return (...args) => {
        clearTimeout(timer);
        timer = setTimeout(() => {
            callback(...args);
        }, waitTime);
    };
}

const debounceHandler = debounce(searchData, 1000);
searchElement.addEventListener('input', debounceHandler);

let url = new URL(window.location.href)
if (url.searchParams.has('wildcard')) {
    // update elements 1st before doing a search later
    wildcardElement.checked = stringToBoolean(url.searchParams.get('wildcard'))
}
if (url.searchParams.has('qp')) {
    body.className = "bodyWithResults"
    searchElement.value = url.searchParams.get('qp')
    searchElement.dispatchEvent(new Event('input'));
}
