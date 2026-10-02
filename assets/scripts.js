const formName = "fileUploadForm";
const inputName = {
    userName: "username",
    userPassword: "userpwd",
    folderPath: "folderpath",
    fileHashes: "filehashes",
    files: "files",
    uploadBtn: "upload_btn",
};
const url = {
    upload: "/upload",
    serverQueueSize: "/queue",
}
const message = {
    selectedFiles: "Selected files:",
    calculatingHashSum: `Calculating hash sums. Please, wait ... <br> File %s of %s. %s.`,
    sendingFiles: "Sending files. Please, wait ...",
    uploadFailed: "Upload failed. More details are available in console.",
    serverIsBusy: "Server is busy. Try again later.",
    successfulUpload: "Successful upload",
};
const error = {
    server: "Server error:",
    userNameIsNotSet: "User name is not set",
    userPasswordIsNotSet: "User password is not set",
    fileHashesAreNotSet: "File hash sums (SHA-256) are not set",
    filesAreNotSet: "Files are not set",
    fileIsEmpty: "File [%s] is empty",
};
const id = {
    status: "status",
    fileRow: "file_row",
    progressbar: "progressbar",
    serverQueueSize: "server_queue_size",
    serverResponseTime: "server_response_time",
};
let lastRequestStatusCode = 0;
let sqsPingMs = 0;
const interval = {
    updateSQS: 5,
}

window.addEventListener("load", () => {
    onBodyLoaded().then(r => {
    });
});

async function onBodyLoaded() {
    await updateSQS();
    setInterval(updateSQS, interval.updateSQS * 1000);
}

async function updateSQS() {
    let divSQS = document.getElementById(id.serverQueueSize);
    let divSRT = document.getElementById(id.serverResponseTime);

    divSQS.textContent = "";
    divSRT.textContent = "";

    const timeStart = performance.now();
    let sqs = await fetchServerQueueSize();
    const timeEnd = performance.now();

    sqsPingMs = timeEnd - timeStart;
    divSQS.textContent = sqs.toString();
    divSRT.textContent = sqsPingMs + " ms";
}

async function onFileInputChange(element) {
    console.log(message.selectedFiles, element.files);
    enableControls(false);

    let hashes = [];
    for (let i = 0; i < element.files.length; i++) {
        let msg = formatString(message.calculatingHashSum, i + 1, element.files.length, element.files[i].name);
        updateStatus(msg);

        let hash = await calculateFileHashSumSha256(element.files[i]);
        hashes.push(hash);
    }

    let form = document.forms[formName];
    let fileHashes = form.elements[inputName.fileHashes];
    fileHashes.value = JSON.stringify(hashes);
    enableControls(true);
    updateStatus('');
}

async function onUploadBtnClick() {
    if (!checkForm()) {
        return false;
    }

    enableControls(false);
    updateStatus(message.sendingFiles);
    let ok = await sendForm();
    enableControls(true);
    if (!ok) {
        if (lastRequestStatusCode === 503) {
            updateStatus(message.serverIsBusy);
        } else {
            updateStatus(message.uploadFailed);
        }
        return false;
    }

    updateStatus(message.successfulUpload);
    return true;
}

function checkForm() {
    let form = document.forms[formName];

    let userName = form.elements[inputName.userName].value;
    let userPassword = form.elements[inputName.userPassword].value;
    let filesInput = form.elements[inputName.files];
    let fileHashes = form.elements[inputName.fileHashes].value;

    if (userName.length < 1) {
        updateStatus(error.userNameIsNotSet);
        return false;
    }
    if (userPassword.length < 1) {
        updateStatus(error.userPasswordIsNotSet);
        return false;
    }
    if (fileHashes.length < 1) {
        updateStatus(error.fileHashesAreNotSet);
        return false;
    }
    if (filesInput.files.length === 0) {
        updateStatus(error.filesAreNotSet);
        return false;
    }
    for (let i = 0; i < filesInput.files.length; i++) {
        if (filesInput.files[i].size === 0) {
            updateStatus(formatString(error.fileIsEmpty, i.toString()));
            return false;
        }
    }

    return true;
}

async function sendForm() {
    return await sendFormUsingXHRAsync();
}

async function sendFormUsingXHRAsync() {
    return new Promise((resolve, reject) => {
        let form = document.forms[formName];
        let formData = new FormData(form);
        let xhr = new XMLHttpRequest();

        xhr.upload.addEventListener('progress', (event) => {
            if (event.lengthComputable) {
                let progress = Math.round((event.loaded / event.total) * 100);
                setProgress(progress);
            }
        });

        xhr.addEventListener('load', () => {
            lastRequestStatusCode = xhr.status;
            if (xhr.status >= 200 && xhr.status < 300) {
                resolve(true);
            } else {
                resolve(false);
            }
        });

        xhr.onerror = function () {
            resolve(false);
        };

        xhr.open('POST', url.upload);
        xhr.send(formData);
    });
}

// N.B.: This function can not show upload progress.
async function sendFormUsingFetch() {
    let form = document.forms[formName];
    let formData = new FormData(form);
    let ri = {
        method: 'POST',
        body: formData
    };
    let response = await fetch(url.upload, ri);

    if (!(response.ok)) {
        console.error(error.server, response.status);
        return false;
    }

    if (response.status !== 200) {
        console.error(error.server, response.status);
        return false;
    }

    return true;
}

async function calculateFileHashSumSha256(file) {
    const chunkSize = 1024 * 1024;
    const chunksCount = Math.ceil(file.size / chunkSize);
    let chunkNum = 1;
    const hasher = sha256.create();
    let offset = 0;
    let progress = 0.0;

    while (offset < file.size) {
        progress = chunkNum / chunksCount;
        setProgress(progress * 100);

        const chunk = file.slice(offset, offset + chunkSize);
        let buffer = await chunk.arrayBuffer();
        hasher.update(buffer);
        buffer = null;
        offset += chunkSize;
        chunkNum++;
    }

    return hasher.hex();
}

function formatString(template, ...args) {
    let index = 0;
    return template.replace(/%s/g, () => args[index++] ?? '%s');
}

function updateStatus(text) {
    let idDiv = document.getElementById(id.status);
    idDiv.innerHTML = text;
}

function enableControls(isEnabled) {
    let form = document.forms[formName];

    let uploadBtn = form.elements[inputName.uploadBtn];
    uploadBtn.disabled = !isEnabled;

    let folderPathInput = form.elements[inputName.folderPath];
    folderPathInput.readOnly = !isEnabled;

    let userPasswordInput = form.elements[inputName.userPassword];
    userPasswordInput.readOnly = !isEnabled;

    let userNameInput = form.elements[inputName.userName];
    userNameInput.readOnly = !isEnabled;

    let fileRow = document.getElementById(id.fileRow);
    if (isEnabled) {
        fileRow.style.display = "table-row";
    } else {
        fileRow.style.display = "none";
    }
}

function setProgress(percent) {
    let pb = document.getElementById(id.progressbar);
    let cp = Math.min(Math.max(percent, 0), 100);
    pb.style.width = cp + "%";
}

async function sleep(ms) {
    await new Promise(resolve => setTimeout(resolve, ms));
}

async function fetchServerQueueSize() {
    try {
        let response = await fetch(url.serverQueueSize);

        if (!response.ok) {
            console.error(error.server, response.status);
            return null;
        }

        let data = await response.json();
        return data.queue;
    } catch (error) {
        console.error(error);
        return null;
    }
}
