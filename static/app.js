let formConfig = {}
let configFromApp = {}
let connected = false
let needRefresh = true

let ws

var headers = document.querySelectorAll('.dropdown-container header');

for (var i = 0; i < headers.length; i++) {
    headers[i].addEventListener('click', openAccordion);
}

function openAccordion(e) {
    var parent = this.parentElement;
    var article = this.nextElementSibling;

    if (!parent.classList.contains('open')) {
        parent.classList.add('open');
        article.style.maxHeight = article.scrollHeight + 'px';
    } else {
        parent.classList.remove('open');
        article.style.maxHeight = '0px';
    }
}

    // open first panel
    /*{
        let parent = headers[0].parentElement;
        let article = headers[0].nextElementSibling;
        if (!parent.classList.contains('open')) {
            parent.classList.add('open');
            article.style.maxHeight = article.scrollHeight + 'px';
        }
    }*/

let loading = false
let sending = false
    
// getStatus();
// window.setInterval(getStatus, 2000);

function setVtsState(state) {
    // console.log("vts " + state)
    const s1 = document.getElementById('vtsStatus').getElementsByTagName('li')
    for (const li of s1) {
        li.classList.remove("good", "bad")
        if (li.id == 'vts_' + state) {
            const status = state == 'finished' ? 'bad' : 'good'
            li.classList.add(status)
        }
    }
}

function setDaState(state) {
    // console.log("da " + state)
    const s2 = document.getElementById('daStatus').getElementsByTagName('li')
    for (const li of s2) {
        li.classList.remove("good", "bad")
        if (li.id == 'da_' + state) {
            const status = state == 'finished' ? 'bad' : 'good'
            li.classList.add(status)
        }
        const link = document.getElementById('authLink')
        if (state == "waiting") {
            if (link.classList.contains('hidden')) {
                link.classList.remove('hidden')
            }
        } else {
            if (!link.classList.contains('hidden')) {
                document.getElementById('authLink').classList.add('hidden')
            }
        }
    }
}

function getStatus() {
    // console.log(loading)
    if (!loading) {
        loading = true
        fetch("/api/status")
            .then((response) => {
                loading = false
                return response.json()
            })
            .then((data) => {
                configFromApp = data.config
                if (!connected) {
                    resetForm();
                }
                setVtsState(data.vtsState);
                setDaState(data.daState);
                document.getElementById('authLink').setAttribute('href', data.authLink);
                connected = true
            })
            .catch((err) => {
                loading = false
                connected = false
                setVtsState('finished')
                setDaState('finished')
            });
    }
}

function validateDiff(diff) {
    return true
}

function getDiff() {
    const vtsAddr = document.getElementById('vtsAddr').value
    
}

function setStringField(c, name) {
    if (c[name] != null) {
        document.getElementById(name).value = c[name]
    }
}

function setNumberField(c, name) {
    if (c[name] != null) {
        document.getElementById(name).value = c[name]
    }
}

function setCheckbox(c, name) {
    if (c[name] != null) {
        document.getElementById(name).checked = configFromApp[name]
    }
}

function resetForm() {
    console.log('form reset')
    if (configFromApp != null) {
        setStringField(configFromApp, "vtsAddr")
        setStringField(configFromApp, "vtsToken")
        setStringField(configFromApp, "customParam")
        setStringField(configFromApp, "paramDescription")
        setNumberField(configFromApp, "stayTime")
        setNumberField(configFromApp, "decayTime")
        setNumberField(configFromApp, "multiplier")
        setCheckbox(configFromApp, "addParam")
        setCheckbox(configFromApp, "logVtsMessages")
    
        setStringField(configFromApp, "daToken")
        // setStringField(configFromApp, "daAppId")
        // setNumberField(configFromApp, "daPort")
        setCheckbox(configFromApp, "logDaMessages")

        setCheckbox(configFromApp, "autoOpenUi")
    }
}

function resetTokens() {
    const config = {"vtsToken": "", "daToken": ""}
    sendConfig(config)
}

function validateForm() {
    return true;
}

function toggleInput(enabled) {
    for (e of document.getElementsByTagName('form')[0].getElementsByTagName('input')) {e.disabled=!enabled}
    for (e of document.getElementsByTagName('form')[0].getElementsByTagName('button')) {e.disabled=!enabled}
}

function handleSubmitForm(event) {
    event.preventDefault();
    if (validateForm()) {
        let newConfig = {}
        newConfig["vtsAddr"] = document.getElementById('vtsAddr').value
        newConfig["vtsToken"] = document.getElementById('vtsToken').value
        newConfig["customParam"] = document.getElementById('customParam').value
        newConfig["paramDescription"] = document.getElementById('paramDescription').value
        newConfig["stayTime"] = parseInt(document.getElementById('stayTime').value)
        let decayTime = parseInt(document.getElementById('decayTime').value)
        if (decayTime != NaN) {
            if (decayTime < 1) {
                decayTime = 1
            }
        }
        newConfig["decayTime"] = decayTime
        newConfig["multiplier"] = parseFloat(document.getElementById('multiplier').value)
        newConfig["addParam"] = document.getElementById('addParam').checked
        newConfig["daToken"] = document.getElementById('daToken').value
        // newConfig["daAppId"] = document.getElementById('daAppId').value
        // newConfig["daPort"] = parseInt(document.getElementById('daPort').value)
        newConfig["autoOpenUi"] = document.getElementById('autoOpenUi').checked
        // console.table(newConfig)
        sendConfig(newConfig)
    } else {
        // highlight errors
    }
}

function sendConfig(config) {
    if (!sending) {
        sending = true
        toggleInput(false)
        fetch("/api/setConfig", {method: "POST", headers: {"Content-Type": "application/json"}, body: JSON.stringify(config)})
            .then((response) => {
                sending = false
                toggleInput(true)
                return response.json()
            })
            .then((data) => {
                // console.table(data.config)
                configFromApp = data.config
                resetForm();
                setVtsState(data.vtsState);
                setDaState(data.daState);
                document.getElementById('authLink').setAttribute('href', data.authLink);
            })
            .catch((err) => {
                console.log(err)
                sending = false
                toggleInput(true)
                setVtsState('finished')
                setDaState('finished')
            });
    }
}

document.querySelector('#formReset').addEventListener('click', resetForm)
document.querySelector('#tokensReset').addEventListener('click', resetTokens)
document.querySelector('form').addEventListener('submit', handleSubmitForm);

function reconnectWs() {
    if (ws) {
        return false;
    }
    ws = new WebSocket("ws://" + document.location.host + "/ws");
    ws.onopen = function(evt) {
        console.log("ws opened");
        connected = true;
        needRefresh = true;
    }
    ws.onclose = function(evt) {
        console.log("ws closed");
        connected = false;
        ws = null;
        setVtsState('finished')
        setDaState('finished')
        window.setTimeout(reconnectWs, 2000)
    }
    ws.onmessage = function(evt) {
        // console.log("UPDATE: " + evt.data);
        const data = JSON.parse(evt.data)
        configFromApp = data.config
        if (needRefresh) {
            resetForm();
        }
        setVtsState(data.vtsState);
        setDaState(data.daState);
        document.getElementById('authLink').setAttribute('href', data.authLink);
        needRefresh = false;
    }
    ws.onerror = function(evt) {
        console.log("ERROR: " + evt.data);
    }
    return false;
}

reconnectWs()