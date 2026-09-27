const formName = "fileUploadForm";

async function onUploadBtnClick() {
    if (!checkForm()) {
        return false;
    }

    let ok = await sendForm();
    if (!ok) {
        return false;
    }

    console.log("Successful upload.");
    return true;
}

function checkForm() {
    let form = document.forms[formName];

    let userName = form.elements['username'].value;
    let userPassword = form.elements['userpwd'].value;
    let file = form.elements['file'];
    let fileHash_sha256 = form.elements['filehash_sha256'].value;

    if (userName.length < 1) {
        console.error("User name is not set");
        return false;
    }
    if (userPassword.length < 1) {
        console.error("User password is not set");
        return false;
    }
    if (fileHash_sha256.length < 1) {
        console.error("File hash sum (SHA-256) is not set");
        return false;
    }
    if (file.files.length !== 1) {
        console.error("File is not set");
        return false;
    }
    if (file.files[0].size === 0) {
        console.error("File is empty");
        return false;
    }

    return true;
}

async function sendForm() {
    let form = document.forms[formName];
    let formData = new FormData(form);
    let uploadBtn = form.elements['upload_btn'];
    uploadBtn.disabled = true;

    let url = '/upload';
    let ri = {
        method: 'POST',
        body: formData
    };
    let response = await fetch(url, ri);

    if (!(response.ok)) {
        console.error('Server error:', response.status);
        uploadBtn.disabled = false;
        return false;
    }

    if (response.status !== 200) {
        console.error('Server error:', response.status);
        uploadBtn.disabled = false;
        return false;
    }

    uploadBtn.disabled = false;
    return true;
}
