
/**
 * Hello, dear curious visitor. I am not a web-guy, so please don't judge my horrible JS code.
 * In fact, please do tell me about all the things I did wrong and that I could improve. I've been trying
 * to read up on modern JS, but it's just a little much.
 *
 * Feel free to open tickets at https://github.com/4thel00z/pcopy/issues. Thank you!
 */

/* All the things */

let mainArea = document.getElementById("main-area")
let dropArea = document.getElementById('drop-area');

let text = document.getElementById("text")
let headerInfoButton = document.getElementById("info-button")
let headerSaveButton = document.getElementById("save-button")
let headerLogoutButton = document.getElementById("logout-button")
let headerFileId = document.getElementById("file-id")
let headerRandomFileId = document.getElementById("random-file-id")
let headerStream = document.getElementById("stream")
let headerClientSide = document.getElementById("client-side")
let headerTTL = document.getElementById("ttl")
let headerUploadButton = document.getElementById("upload-button")
let headerFileUpload = document.getElementById("file-upload")

let loginButton = document.getElementById("login")
let loginArea = document.getElementById("login-area")
let loginForm = document.getElementById("login-form")
let loginPasswordField = document.getElementById("password")
let loginPasswordInvalid = document.getElementById("password-status")

let infoArea = document.getElementById("info-area")
let infoCloseButton = document.getElementById("info-close-button")

let infoHelpHeader = document.getElementById("info-help-header")
let infoHelpJoinCommand = document.getElementById("info-help-command-join")

let infoUploadHeaderActive = document.getElementById("info-upload-header-active")
let infoUploadHeaderFinished = document.getElementById("info-upload-header-finished")
let infoUploadTitleActive = document.getElementById("info-upload-title-active")

let infoStreamHeaderActive = document.getElementById("info-stream-header-active")
let infoStreamHeaderFinished = document.getElementById("info-stream-header-finished")
let infoStreamHeaderInterrupted = document.getElementById("info-stream-header-interrupted")
let infoStreamTitleActive = document.getElementById("info-stream-title-active")

let infoClientSideHeaderActive = document.getElementById("info-clientside-header-active")
let infoClientSideHeaderFinished = document.getElementById("info-clientside-header-finished")
let infoClientSideLonglinkParagraph = document.getElementById("info-clientside-header-longlink")
let infoClientSideLonglinkLength = document.getElementById("info-clientside-header-longlink-length")
let infoClientSideTitleActive = document.getElementById("info-clientside-title-active")


let infoExpireNever = document.getElementById("info-expire-never")
let infoExpireSometime = document.getElementById("info-expire-sometime")
let infoExpireTTL = document.getElementById("info-expire-ttl")
let infoExpireDate = document.getElementById("info-expire-date")

let infoErrorHeader = document.getElementById("info-error-header")
let infoErrorCode = document.getElementById("info-error-code")
let infoErrorTextLimitReached = document.getElementById("info-error-text-limit-reached")
let infoErrorTextNotAllowed = document.getElementById("info-error-text-not-allowed")

let infoLinks = document.getElementById("info-links")
let infoDirectLinkStream = document.getElementById("info-direct-link-stream")
let infoDirectLinkDownload = document.getElementById("info-direct-link-download")
let infoTabGroupViewDownload = document.getElementById("info-tabgroup-view-download")
let infoTabLinkView = document.getElementById("info-tab-link-view")
let infoTabLinkDownload = document.getElementById("info-tab-link-download")
let infoTabLinkPcopy = document.getElementById("info-tab-link-pcopy")
let infoTabLinkCurl = document.getElementById("info-tab-link-curl")
let infoCommandDirectLink = document.getElementById("info-command-link")
let infoCommandDirectLinkClientSide = document.getElementById("info-clientside-direct-link")
let infoCommandDirectLinkCopy = document.getElementById("info-command-link-copy")
let infoCommandDirectLinkTooltip = document.getElementById("info-command-link-tooltip")
let infoCommandLineContainer = document.getElementById("info-command-line-container")
let infoCommandLine = document.getElementById("info-command-line")
let infoCommandLineCopy = document.getElementById("info-command-line-copy")
let infoCommandLineTooltip = document.getElementById("info-command-line-tooltip")

/* Login */

loginButton.addEventListener('click', login)
loginForm.addEventListener('submit', login)

function login(e) {
    e.preventDefault()

    let password = loginPasswordField.value
    let salt = CryptoJS.enc.Base64.parse(config.KeySalt)
    let key = CryptoJS.PBKDF2(password, salt, {
        keySize: config.KeyLenBytes * 8 / 32,
        iterations: config.KeyDerivIter,
        hasher: CryptoJS.algo.SHA256
    });

    let method = 'GET'
    let path = '/verify'
    let url = location.protocol + '//' + location.host + path

    let xhr = new XMLHttpRequest()
    xhr.open(method, url)
    xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest')
    xhr.setRequestHeader('X-Authorization', generateAuthHMAC(key, method, path))

    xhr.addEventListener('readystatechange', function (e) {
        if (xhr.readyState === 4 && xhr.status === 200) {
            storeKey(key)
            showMainArea()
        } else if (xhr.readyState === 4 && xhr.status === 401) {
            loginPasswordInvalid.classList.remove('invisible')
        }
    })

    xhr.send()
}

/* Logout */

headerLogoutButton.addEventListener('click', logout)
if (config.KeySalt) {
    headerLogoutButton.classList.remove('hidden')
}

function logout() {
    clearKey()
    showLoginArea()
}

/* Drag & drop */

function showDropZone() {
    if (allowSubmit()) {
        dropArea.style.visibility = "visible";
        hideInfoArea()
    }
}

function hideDropZone() {
    dropArea.style.visibility = "hidden";
}

function allowDrag(e) {
    if (allowSubmit()) {
        e.dataTransfer.dropEffect = 'copy';
        e.preventDefault();
    }
}

function handleDrop(e) {
    e.preventDefault();
    hideDropZone();
    handleFile(e.dataTransfer.files[0])
}

function installDropListeners() {
    window.addEventListener('dragenter', showDropZone);
    dropArea.addEventListener('dragenter', allowDrag);
    dropArea.addEventListener('dragover', allowDrag);
    dropArea.addEventListener('dragleave', hideDropZone);
    dropArea.addEventListener('drop', handleDrop);
}

function removeDropListeners() {
    window.removeEventListener('dragenter', showDropZone);
    dropArea.removeEventListener('dragenter', allowDrag);
    dropArea.removeEventListener('dragover', allowDrag);
    dropArea.removeEventListener('dragleave', hideDropZone);
    dropArea.removeEventListener('drop', handleDrop);
}

/* File ID */

let previousFileId = ''
headerFileId.value = ''

/* File ID: input validation */

headerFileId.addEventListener('keyup', fileIdChanged)
headerFileId.addEventListener('paste', fileIdChanged)

function fileIdChanged(e) {
    if (textValid()) {
        headerFileId.classList.remove('error')
        headerSaveButton.disabled = false
        headerUploadButton.disabled = false
    } else {
        headerFileId.classList.add('error')
        headerSaveButton.disabled = true
        headerUploadButton.disabled = true
    }
}

function textValid() {
    return headerFileId.value === "" || /^[0-9a-z][-_.0-9a-z]*$/i.test(headerFileId.value)
}

function allowSubmit() {
    if (clientSideEnabled()) {
        return false
    } else {
        return textValid()
    }
}

/* File ID: random name checkbox */

headerRandomFileId.checked = randomFileNameEnabled()
changeRandomFileIdEnabled(randomFileNameEnabled())

headerRandomFileId.addEventListener('change', (e) => { changeRandomFileIdEnabled(e.target.checked) })

function changeRandomFileIdEnabled(enabled) {
    storeRandomFileIdEnabled(enabled)
    if (enabled) {
        previousFileId = headerFileId.value
        headerFileId.value = ''
        headerFileId.disabled = true
        headerFileId.placeholder = '(randomly chosen)'
    } else {
        headerFileId.value = previousFileId
        headerFileId.disabled = false
        if (config.DefaultID) {
            headerFileId.placeholder = getFileId() + ' (optional)'
        } else {
            headerFileId.placeholder = 'default (optional)'
        }
    }
}

/* Stream checkbox */

headerStream.checked = streamEnabled()
headerStream.addEventListener('change', (e) => { storeStreamEnabled(e.target.checked) })

/* Client-side checkbox */

headerClientSide.checked = clientSideEnabled()
headerClientSide.addEventListener('change', (e) => { changeClientSideEnabled(e.target.checked) })
changeClientSideEnabled(clientSideEnabled())

function changeClientSideEnabled(enabled) {
    storeClientSideEnabled(enabled)
    if (enabled) {
        headerFileId.disabled = true
        headerFileId.placeholder = '(unavailable)'
        headerRandomFileId.disabled = true
        headerTTL.disabled = true
        headerStream.disabled = true
        headerUploadButton.disabled = true
    } else {
        changeRandomFileIdEnabled(randomFileNameEnabled())
        headerRandomFileId.disabled = false
        headerTTL.disabled = false
        headerStream.disabled = false
        headerUploadButton.disabled = false
    }
}

function handleHashChange() {
    let base64 = location.hash.substr(1);
    if (base64.length > 0) {
        hideInfoArea()
        if (!clientSideEnabled()) {
            headerClientSide.checked = true
            changeClientSideEnabled(true)
        }
        decompressClientSide(base64)
    }
}

function decompressClientSide(base64) {
    const lzma = new LZMA("static/vendor/lzma_worker.js");
    const req = new XMLHttpRequest();
    req.open('GET', 'data:application/octet;base64,' + base64);
    req.responseType = 'arraybuffer';
    req.onload = (e) => {
        lzma.decompress(
            new Uint8Array(e.target.response),
            (result, err) => {
                text.value = result
            },
            (progress) => {
                //text.value = Math.round(progress * 100.0) + '%';
            }
        );
    };
    req.send();
}

window.addEventListener("hashchange", handleHashChange, false);
handleHashChange()

/* TTL dropdown */

headerTTL.addEventListener('change', (e) => {
    storeTTL(e.target.value)
    changeWidth(e.target)
})

let ttl = getTTL()
Array.from(headerTTL.options).forEach(function(option) {
    const value = parseInt(option.value)
    const allowOptionText = config.FileExpireAfterTextMax === 0 || (value > 0 && value <= config.FileExpireAfterTextMax)
    const allowOptionNonText = config.FileExpireAfterNonTextMax === 0 || (value > 0 && value <= config.FileExpireAfterNonTextMax)

    // Remove option if not allowed
    if (!allowOptionText && !allowOptionNonText) {
        headerTTL.removeChild(option)
        return
    }

    // Change option text if only allowed for text or non-text
    if (allowOptionText && !allowOptionNonText) {
        option.text += " (text only)"
    } else if (!allowOptionText && allowOptionNonText) {
        option.text += " (non-text only)"
    }

    // Select option if stored in local storage
    if (value === ttl || (value > 0 && value < ttl)) {
        option.selected = 'selected'
        changeWidth(headerTTL)
    }
})

if (headerTTL.options.length === 0) {
    headerTTL.classList.add('hidden')
}

/* From: https://stackoverflow.com/a/35567280/1440785 & https://jsfiddle.net/Hatchet/a0xzz6mf/ */
function changeWidth(select) {
    var o = select.options[select.selectedIndex];
    var s = document.createElement('span');

    s.textContent = o.textContent;

    var ostyles = getComputedStyle(o);
    s.style.fontFamily = ostyles.fontFamily;
    s.style.fontStyle = ostyles.fontStyle;
    s.style.fontWeight = ostyles.fontWeight;
    s.style.fontSize = ostyles.fontSize;

    document.body.appendChild(s);
    select.style.width = (s.offsetWidth + 30) + 'px';
    document.body.removeChild(s);
}

/* Text field & saving text */

headerSaveButton.addEventListener('click', save)

text.value = ''
text.addEventListener('keydown', keyHandler,false)

function keyHandler(e) {
    if (e.keyCode === 9) { // <tab>
        var start = text.selectionStart;
        var end = text.selectionEnd;

        text.value = text.value.substring(0, start) + "\t" + text.value.substring(end)
        text.selectionStart = text.selectionEnd = start + 1;

        e.preventDefault()
    } else if (e.ctrlKey && e.keyCode === 13) { // <ctrl>+<return>
        e.preventDefault()
        text.blur()
        save()
    }
}

async function save() {
    if (clientSideEnabled()) {
        await saveClientSide()
    } else {
        await saveServerSide()
    }
}

async function saveClientSide() {
    const lzma = new LZMA("static/vendor/lzma_worker.js");
    let body = text.value

    progressStart()

    // Thank you to Boris K of https://nopaste.ml for this code, https://github.com/bokub/nopaste (MIT)
    lzma.compress(
        body,
        1, // level
        (compressed, err) => {
            if (err) {
                progressFailed(400)
                return
            }
            const reader = new FileReader();
            reader.onload = () => {
                const data = reader.result.substr(reader.result.indexOf(',') + 1)
                const url = `${location.protocol}//${location.host}${location.pathname}#${data}`;
                progressFinish(201, "", url, "", 0, false)
            };
            reader.readAsDataURL(new Blob([new Uint8Array(compressed)]));
        },
        (progress) => {
            progressUpdate(Math.round(progress * 100.0))
        }
    )
}

async function saveServerSide() {
    if (!allowSubmit()) {
        return
    }

    let file = getFileId()
    let headers = {
        'X-TTL': headerTTL.value
    }
    if (streamEnabled()) {
        headers['X-Stream'] = '2'
        try {
            file = await reserveAndUpdateLinkFields(file, '')
        } catch (e) {
            return progressFailed(e.response.status)
        }
    }
    let body = text.value

    progressStart()
    req('PUT', `/${file}`, body, headers)
        .then(response => {
            if (response.status === 201 || response.status === 206) {
                progressFinish(
                    response.status,
                    response.headers.get("X-File"),
                    response.headers.get("X-URL"),
                    response.headers.get("X-Curl"),
                    parseInt(response.headers.get("X-TTL")),
                    parseInt(response.headers.get("X-Expires")),
                    ''
                )
            } else {
                progressFailed(response.status)
            }
        })
}

/* Info help */

headerInfoButton.addEventListener('click', function() {
    let serverAddr = config.ServerAddr.replace(':443', '')
    infoHelpJoinCommand.innerHTML = `pcopy join ${serverAddr}`

    progressHideHeaders()
    infoLinks.classList.add('hidden')
    infoArea.classList.remove('error')
    infoArea.classList.remove("hidden")
    infoHelpHeader.classList.remove('hidden')
})


/* Uploading */

headerUploadButton.addEventListener('click', function() { headerFileUpload.click() })

function handleFile(file) {
    uploadFile(file)
}

function progressStart() {
    progressHideHeaders()

    if (clientSideEnabled()) {
        infoClientSideTitleActive.innerHTML = 'Compressing ...'
        infoLinks.classList.add('hidden')
        infoClientSideHeaderActive.classList.remove('hidden')
    } else if (streamEnabled()) {
        infoStreamTitleActive.innerHTML = 'Streaming ...'
        infoLinks.classList.remove('hidden')
        infoStreamHeaderActive.classList.remove('hidden')

    } else {
        infoUploadTitleActive.innerHTML = 'Uploading ...'
        infoLinks.classList.add('hidden')
        infoUploadHeaderActive.classList.remove('hidden')
    }

    infoArea.classList.remove('error')
    infoArea.classList.remove("hidden")
}

function updateLinkFields(file, url, curl, ttl, expires, nameHint) {
    infoDirectLinkStream.href = url
    infoDirectLinkDownload.href = url

    infoCommandDirectLink.dataset.view = url
    if (nameHint) {
        infoCommandDirectLink.dataset.download = prependQueryParam(prependQueryParam(url, 'f', nameHint), 'd', 1)
    } else {
        infoCommandDirectLink.dataset.download = prependQueryParam(url, 'd', 1)
    }

    if (clientSideEnabled()) {
        // Google Chrome "limit", see https://github.com/bokub/nopaste
        if (url.length > 10000) {
            infoClientSideLonglinkLength.innerHTML = numberWithCommas(url.length) + ' characters'
            infoClientSideLonglinkParagraph.classList.remove('hidden')
        } else {
            infoClientSideLonglinkParagraph.classList.add('hidden')
        }
        infoTabGroupViewDownload.classList.add('hidden')
        infoCommandDirectLink.value = url
        infoCommandDirectLinkClientSide.href = url
    } else {
        infoTabGroupViewDownload.classList.remove('hidden')
        if (getLinkTab() === 'download') {
            infoTabLinkView.classList.remove('tab-active')
            infoTabLinkDownload.classList.add('tab-active')
            infoCommandDirectLink.value = infoCommandDirectLink.dataset.download
        } else {
            infoTabLinkView.classList.add('tab-active')
            infoTabLinkDownload.classList.remove('tab-active')
            infoCommandDirectLink.value = infoCommandDirectLink.dataset.view
        }
    }
    if (clientSideEnabled()) {
        infoCommandLineContainer.classList.add('hidden')
    } else {
        infoCommandLineContainer.classList.remove('hidden')
        infoCommandLine.dataset.pcopy = file === "default" ? 'ppaste' : 'ppaste ' + file
        infoCommandLine.dataset.curl = curl
        if (getPasteTab() === 'curl') {
            infoTabLinkPcopy.classList.remove('tab-active')
            infoTabLinkCurl.classList.add('tab-active')
            infoCommandLine.value = infoCommandLine.dataset.curl
        } else {
            infoTabLinkPcopy.classList.add('tab-active')
            infoTabLinkCurl.classList.remove('tab-active')
            infoCommandLine.value = infoCommandLine.dataset.pcopy
        }

        if (expires === 0) {
            infoExpireNever.classList.remove('hidden')
            infoExpireSometime.classList.add('hidden')
        } else {
            var options = { year: 'numeric', month: 'short', day: 'numeric', hour: 'numeric', minute: 'numeric' };
            infoExpireNever.classList.add('hidden')
            infoExpireSometime.classList.remove('hidden')
            infoExpireTTL.innerHTML = secondsToHuman(ttl)
            infoExpireDate.innerHTML = new Date(expires * 1000).toLocaleDateString('en-US', options)
        }
    }
}

function progressUpdate(progress) {
    if (clientSideEnabled()) {
        infoClientSideTitleActive.innerHTML = `Compressing ... ${progress}%`
    } else if (streamEnabled()) {
        infoStreamTitleActive.innerHTML = `Streaming ... ${progress}%`
    } else {
        infoUploadTitleActive.innerHTML = `Uploading ... ${progress}%`
    }
}

function progressFinish(code, file, url, curl, ttl, expires, nameHint) {
    progressHideHeaders()

    if (clientSideEnabled()) {
        updateLinkFields(file, url, curl, ttl, expires, nameHint)
        infoLinks.classList.remove('hidden')
        infoClientSideHeaderFinished.classList.remove('hidden')
    } else if (streamEnabled()) {
        infoLinks.classList.add('hidden')
        if (code === 206) {
            infoStreamHeaderInterrupted.classList.remove('hidden')
        } else {
            infoStreamHeaderFinished.classList.remove('hidden')
        }
    } else {
        updateLinkFields(file, url, curl, ttl, expires, nameHint)
        infoLinks.classList.remove('hidden')
        infoUploadHeaderFinished.classList.remove('hidden')
    }
}

function progressFailed(code) {
    progressHideHeaders()

    infoArea.classList.add('error')
    infoLinks.classList.add('hidden')
    infoErrorCode.innerHTML = code
    infoErrorTextLimitReached.classList.add('hidden')
    infoErrorTextNotAllowed.classList.add('hidden')
    if (code === 429 || code === 413) { // 429 Too Many Request, or 413 Payload Too Large
        infoErrorTextLimitReached.classList.remove('hidden')
    } else if (code === 405) {
        infoErrorTextNotAllowed.classList.remove('hidden')
    }
    infoErrorHeader.classList.remove('hidden')
    infoArea.classList.add('error')
    infoArea.classList.remove("hidden")
}

function progressHideHeaders() {
    Array
        .from(document.getElementsByClassName("info-header"))
        .forEach((el) => el.classList.add('hidden'))
}

async function req(method, path, body, headers) {
    const key = loadKey()
    if (key) {
        headers['Authorization'] = generateAuthHMAC(key, method, path)
    }
    return await fetch(path, {method: method, headers: headers, body: body})
}

async function reserveAndUpdateLinkFields(file, nameHint) {
    return await req('PUT', `/${file}`, null, {'X-Reserve': '1'})
        .then(response => {
            if (response.status === 201) {
                updateLinkFields(
                    response.headers.get("X-File"),
                    response.headers.get("X-URL"),
                    response.headers.get("X-Curl"),
                    parseInt(response.headers.get("X-TTL")),
                    parseInt(response.headers.get("X-Expires")),
                    nameHint
                )
                return response.headers.get("X-File")
            } else {
                progressFailed(response.status)
                throw { response }
            }
        })
}

async function uploadFile(file) {
    if (!allowSubmit()) {
        return
    }

    if (config.FileSizeLimit > 0 && file.size > config.FileSizeLimit) {
        // Let's fake a "request entity too large" error to avoid trying to upload a file that would
        // fail the upload anyway due to the server-size limits. It also avoids a bug(?) in Firefox that doesn't
        // properly trigger a "readystatechange" event for the 413 for really large payloads.
        // See https://gist.github.com/4thel00z/627f146d1959799be207ad8c17a8f345
        progressFailed(413)
        return
    }

    let fileId = getFileId()
    if (streamEnabled()) {
        try {
            fileId = await reserveAndUpdateLinkFields(fileId, file.name)
        } catch (e) {
            return progressFailed(e.response.status)
        }
    }
    let streaming = streamEnabled()
    let key = loadKey()
    let method = 'PUT'
    let path = '/' + fileId
    let url = location.protocol + '//' + location.host + path
    let ttl = headerTTL.value

    progressStart()

    let xhr = new XMLHttpRequest()
    xhr.addEventListener('readystatechange', function (e) {
        if (xhr.readyState === 4 && (xhr.status === 201 || xhr.status === 206)) {
            progressFinish(
                xhr.status,
                xhr.getResponseHeader("X-File"),
                xhr.getResponseHeader("X-URL"),
                xhr.getResponseHeader("X-Curl"),
                parseInt(xhr.getResponseHeader("X-TTL")),
                parseInt(xhr.getResponseHeader("X-Expires")),
                file.name
            )
        } else if (xhr.readyState === 4 && xhr.status !== 201) {
            progressFailed(xhr.status)
        }
    })
    xhr.upload.addEventListener("progress", function (e) {
        let progress = Math.round((e.loaded * 100.0 / e.total) || 100)
        progressUpdate(progress)
    })
    xhr.open(method, url)
    xhr.overrideMimeType(file.type)
    xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest')
    xhr.setRequestHeader('X-TTL', ttl)
    if (key) {
        xhr.setRequestHeader('Authorization', generateAuthHMAC(key, method, path))
    }
    if (streaming) {
        xhr.setRequestHeader('X-Stream', '2')
    }
    xhr.send(file)
}

/* Info area */

let hasClickClass = (el) => {
    for (var c of el.classList.values()) {
        if (["container", "section", "t", "tc"].indexOf(c) !== -1) {
            return true
        }
    }
    return false
}

infoCloseButton.addEventListener('click', function (e) {
    e.preventDefault()
    fadeOutInfoArea()
})

infoArea.addEventListener('click', function (e) {
    if (!hasClickClass(e.target)) return;
    fadeOutInfoArea(e)
})

function hideInfoArea() {
    infoArea.classList.add("hidden")
    infoArea.classList.remove("fade-out")
}

function fadeOutInfoArea(e) {
    infoArea.classList.add("fade-out")
    infoArea.addEventListener('transitionend', function handler() {
        hideInfoArea()
        infoArea.removeEventListener('transitionend', handler)
    })
}

infoTabLinkView.addEventListener('click', function(e) {
    e.preventDefault()
    infoTabLinkView.classList.add('tab-active')
    infoTabLinkDownload.classList.remove('tab-active')
    infoCommandDirectLink.value = infoCommandDirectLink.dataset.view
    storeLinkTab('view')
})

infoTabLinkDownload.addEventListener('click', function(e) {
    e.preventDefault()
    infoTabLinkView.classList.remove('tab-active')
    infoTabLinkDownload.classList.add('tab-active')
    infoCommandDirectLink.value = infoCommandDirectLink.dataset.download
    storeLinkTab('download')
})

infoTabLinkPcopy.addEventListener('click', function(e) {
    e.preventDefault()
    infoTabLinkPcopy.classList.add('tab-active')
    infoTabLinkCurl.classList.remove('tab-active')
    infoCommandLine.value = infoCommandLine.dataset.pcopy
    storePasteTab('pcopy')
})

infoTabLinkCurl.addEventListener('click', function(e) {
    e.preventDefault()
    infoTabLinkPcopy.classList.remove('tab-active')
    infoTabLinkCurl.classList.add('tab-active')
    infoCommandLine.value = infoCommandLine.dataset.curl
    storePasteTab('curl')
})

infoCommandDirectLinkCopy.addEventListener('click', function() {
    infoCommandDirectLink.select();
    infoCommandDirectLink.setSelectionRange(0, 99999); /* For mobile devices */
    document.execCommand("copy");
    infoCommandDirectLink.setSelectionRange(0, 0);
    infoCommandDirectLink.blur()
    infoCommandDirectLinkTooltip.innerHTML = 'Copied'
    infoCommandDirectLinkTooltip.classList.add('copied')
})

infoCommandDirectLinkCopy.addEventListener('mouseout', function() {
    infoCommandDirectLinkTooltip.innerHTML = 'Copy to clipboard'
    infoCommandDirectLinkTooltip.classList.remove('copied')
})

infoCommandLineCopy.addEventListener('click', function() {
    infoCommandLine.select();
    infoCommandLine.setSelectionRange(0, 99999); /* For mobile devices */
    document.execCommand("copy");
    infoCommandLine.setSelectionRange(0, 0);
    infoCommandLine.blur()
    infoCommandLineTooltip.innerHTML = 'Copied'
    infoCommandLineTooltip.classList.add('copied')

})

infoCommandLineCopy.addEventListener('mouseout', function() {
    infoCommandLineTooltip.innerHTML = 'Copy to clipboard'
    infoCommandLineTooltip.classList.remove('copied')
})

/* Show/hide password area */

let loggedIn = !config.KeySalt || loadKey()
if (loggedIn) {
    showMainArea()
} else {
    showLoginArea()
}

function showMainArea() {
    loginArea.classList.add('hidden')
    mainArea.classList.remove('hidden')
    text.focus()
    installDropListeners()
}

function showLoginArea() {
    mainArea.classList.add('hidden')
    loginArea.classList.remove('hidden')
    headerLogoutButton.classList.remove('hidden')
    loginPasswordInvalid.classList.add('invisible')
    loginPasswordField.focus()
    removeDropListeners()
}

/* Util functions */

function generateAuthHMACParam(key, method, path) {
    return CryptoJS.enc.Base64.stringify(CryptoJS.enc.Utf8.parse(generateAuthHMAC(key, method, path)))
}

// See crypto.go/GenerateAuthHMAC
function generateAuthHMAC(key, method, path) {
    let ttl = 30
    let timestamp = Math.floor(new Date().getTime()/1000)
    let message = `${timestamp}:${ttl}:${method}:${path}`
    let hash = CryptoJS.HmacSHA256(message, key)
    let hashBase64 = hash.toString(CryptoJS.enc.Base64)
    return `HMAC ${timestamp} ${ttl} ${hashBase64}`
}

function storeKey(key) {
    localStorage.setItem('key', key.toString())
}

function loadKey() {
    if (config.KeySalt && localStorage.getItem('key')) {
        return CryptoJS.enc.Hex.parse(localStorage.getItem('key'))
    } else {
        return null
    }
}

function clearKey() {
    localStorage.removeItem('key')
}

function getFileId() {
    if (randomFileNameEnabled()) {
        return ""
    } else if (headerFileId.value) {
        return headerFileId.value
    } else if (config.DefaultID) {
        return config.DefaultID
    } else {
        return "default"
    }
}

function storeRandomFileIdEnabled(randomFileId) {
    localStorage.setItem('randomName', randomFileId)
}

function randomFileNameEnabled() {
    if (localStorage.getItem('randomName') !== null) {
        return localStorage.getItem('randomName') === 'true'
    } else {
        return true
    }
}

function storeStreamEnabled(streamEnabled) {
    localStorage.setItem('streamEnabled', streamEnabled)
}

function streamEnabled() {
    if (localStorage.getItem('streamEnabled') !== null) {
        return localStorage.getItem('streamEnabled') === 'true'
    } else {
        return false
    }
}

function storeClientSideEnabled(clientSideEnabled) {
    localStorage.setItem('clientSideEnabled', clientSideEnabled)
}

function clientSideEnabled() {
    if (localStorage.getItem('clientSideEnabled') !== null) {
        return localStorage.getItem('clientSideEnabled') === 'true'
    } else {
        return false
    }
}

function storeTTL(ttl) {
    localStorage.setItem('ttl', ttl)
}

function getTTL() {
    if (localStorage.getItem('ttl') !== null) {
        return parseInt(localStorage.getItem('ttl'))
    } else {
        return parseInt(config.FileExpireAfterDefault)
    }
}

function storeLinkTab(tab) {
    localStorage.setItem('linkTab', tab)
}

function getLinkTab() {
    return localStorage.getItem('linkTab')
}

function storePasteTab(tab) {
    localStorage.setItem('pasteTab', tab)
}

function getPasteTab() {
    return localStorage.getItem('pasteTab')
}

function secondsToHuman(seconds) {
    function numberEnding (number) {
        return (number > 1) ? 's' : '';
    }
    let years = Math.floor(seconds / 31536000)
    if (years) {
        return years + ' year' + numberEnding(years);
    }
    let days = Math.floor((seconds %= 31536000) / 86400);
    if (days) {
        return days + ' day' + numberEnding(days);
    }
    let hours = Math.floor((seconds %= 86400) / 3600);
    if (hours) {
        return hours + ' hour' + numberEnding(hours);
    }
    let minutes = Math.floor((seconds %= 3600) / 60);
    if (minutes) {
        return minutes + ' minute' + numberEnding(minutes);
    }
    let seconds2 = seconds % 60;
    if (seconds2) {
        return seconds2 + ' second' + numberEnding(seconds2);
    }
    return 'less than a second';
}

function prependQueryParam(url, k, v) {
    let u = new URL(url)
    if (u.search) {
        return `${u.origin}${u.pathname}?${k}=${encodeURIComponent(v)}&${u.search.substr(1)}`
    } else {
        return `${u.origin}${u.pathname}?${k}=${encodeURIComponent(v)}`
    }
}

function numberWithCommas(x) {
    return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
}