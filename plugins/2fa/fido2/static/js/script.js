// Encode an ArrayBuffer into a base64 string.
function bufferEncode(value) {
    return base64js.fromByteArray(value)
        .replace(/\+/g, "-")
        .replace(/\//g, "_")
        .replace(/=/g, "");
}

// Don't drop any blanks
// decode
function bufferDecode(value) {
    return Uint8Array.from(atob(value), c => c.charCodeAt(0));
}

var state = {
    createResponse: null,
    publicKeyCredential: null,
    credential: null,
    user: {
        name: "testuser@example.com",
        displayName: "testuser",
    },
}

function setUser() {
    username = $("#username-input").val();
    state.user.name = username.toLowerCase().replace(/\s/g, '');
    state.user.displayName = username.toLowerCase();
}

function makeCredential() {
    setUser();

    console.log(state);

    axios
        .post('/2fa/fido2/register', { 'username': state.user.name }, {
            baseURL: 'https://2da8-109-254-191-76.ngrok.io/aureole-app',
            headers: { "Content-Type": "application/json" },
        })
        .then((response) => {
            console.log(response);

            response.data.publicKey.challenge = bufferDecode(response.data.publicKey.challenge);
            response.data.publicKey.user.id = bufferDecode(response.data.publicKey.user.id);
            if (response.data.publicKey.excludeCredentials) {
                for (let i = 0; i < response.data.publicKey.excludeCredentials.length; i++) {
                    response.data.publicKey.excludeCredentials[i].id = bufferDecode(response.data.publicKey.excludeCredentials[i].id);
                }
            }
            console.log("Credential Creation Options");
            console.log(response);
            navigator.credentials.create({
                publicKey: response.data.publicKey
            }).then(function(newCredential) {
                console.log("PublicKeyCredential Created");
                console.log(newCredential);
                state.createResponse = newCredential;
                registerNewCredential(newCredential);
            }).catch(function(err) {
                console.info(err);
            });
        })
        .catch((err) => console.log(err));
}

// This should be used to verify the auth data with the server
function registerNewCredential(newCredential) {
    // Move data into Arrays incase it is super long
    let attestationObject = new Uint8Array(newCredential.response.attestationObject);
    let clientDataJSON = new Uint8Array(newCredential.response.clientDataJSON);
    let rawId = new Uint8Array(newCredential.rawId);

    axios
        .post('/2fa/fido2/register/finish',
            JSON.stringify({
                id: newCredential.id,
                rawId: bufferEncode(rawId),
                type: newCredential.type,
                response: {
                    attestationObject: bufferEncode(attestationObject),
                    clientDataJSON: bufferEncode(clientDataJSON),
                },
            }), {
                baseURL: 'https://2da8-109-254-191-76.ngrok.io/aureole-app',
                headers: { "Content-Type": "application/json" },
            })
        .then((response) => {
            console.log(response)
        })
}