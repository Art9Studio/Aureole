const aureoleURL = "http://localhost:3000/aureole-app";
const getEnabledAuthnURL = "/authn";
const getRedirectURL = "/ui/redirect-url"
const getJWTKeysURL = "/ui/jwt-storage-keys"

const pwbasedRegisterURL = "/password-based/register";
const pwbasedLoginURL = "/password-based/login";
const emailSendLinkURL = "/email/send";
const phoneLoginURL = "/phone/login";
const phoneSendOTPUrl = "/phone/send";
const googleLoginURL = '/google';
const facebookLoginURL = '/facebook';
const vkLoginURL = '/vk';
const appleLoginURL = '/apple';

const googleAuthStartURL = '/2fa/google-authenticator/start';
const googleAuthVerifyURL = '/2fa/google-authenticator/verify';
const smsStartURL = '/2fa/sms/start';
const smsVerifyURL = '/2fa/sms/verify';
const fido2StartURL = '/2fa/fido2/start';
const fido2VerifyURL = '/2fa/fido2/verify';

const expectedProviders = ["google", "vk", "apple", "facebook"];

function sendPostRequest(url, body, successCallback = (resp) => console.log(resp), errorCallback = (err) => console.log(err.message)) {
    axios
        .post(url, body, {
            baseURL: aureoleURL,
            headers: { "Content-Type": "application/json" },
        })
        .then(successCallback)
        .catch(errorCallback);
}

function successAuthnCallback(response) {
    console.log(response);

    if (response.status == 202) {
        app.enabled2FA = response.data['2fa'];
        localStorage.setItem('token', response.data.token);
    } else {
        localStorage.setItem(app.jwtStorageKeys.access, response.headers['X-Access']);
        localStorage.setItem(app.jwtStorageKeys.refresh, response.data['refresh']);
        window.location.replace(app.redirectURL);
    }
}


function success2FAVerifyCallback(response) {
    console.log(response);
    localStorage.setItem(app.jwtStorageKeys.access, response.headers['X-Access']);
    localStorage.setItem(app.jwtStorageKeys.refresh, response.data['refresh']);
    window.location.replace(app.redirectURL);
}

function bufferEncode(value) {
    return base64js.fromByteArray(value)
        .replace(/\+/g, "-")
        .replace(/\//g, "_")
        .replace(/=/g, "");
}

function bufferDecode(value) {
    return Uint8Array.from(atob(value), c => c.charCodeAt(0));
}

app = new Vue({
    el: "#main",
    data() {
        return {
            enabledAuthn: [],
            enabled2FA: [],

            loading: true,
            errored: false,

            googleLoginURL: aureoleURL + googleLoginURL,
            facebookLoginURL: aureoleURL + facebookLoginURL,
            vkLoginURL: aureoleURL + vkLoginURL,
            appleLoginURL: aureoleURL + appleLoginURL,

            redirectURL: '',
            jwtStorageKeys: '',

            pwbasedForm: {
                username: "",
                email: "",
                phone: "",
                password: "",
            },
            emailForm: {
                email: "",
            },
            phoneForm: {
                phone: "",
            },
            phoneOTPForm: {
                otp: "",
            },

            google2FAForm: {
                otp: "",
            },
            sms2FAForm: {
                phone: "",
                otp: "",
            }
        };
    },
    methods: {
        submitPWBasedRegisterForm() {
            sendPostRequest(pwbasedRegisterURL, this.pwbasedForm);
        },
        submitPWBasedLoginForm() {
            sendPostRequest(pwbasedLoginURL, this.pwbasedForm, successAuthnCallback);
        },
        submitEmailForm() {
            sendPostRequest(emailSendLinkURL, this.emailForm, successAuthnCallback);
        },
        submitPhoneForm() {
            sendPostRequest(phoneSendOTPUrl, this.phoneForm, (response) => {
                console.log(response);
                localStorage.setItem("token", response.data.token);
            });
        },
        submitPhoneOTPForm() {
            sendPostRequest(phoneLoginURL, { 'otp': this.phoneOTPForm.otp, 'token': localStorage.getItem("token") }, successAuthnCallback);
        },
        handle2FAChoice(e) {
            switch (e.target.value) {
                case "google_authenticator":
                    axios
                        .post(googleAuthStartURL, { 'token': localStorage.getItem('token') }, {
                            baseURL: aureoleURL,
                            headers: { "Content-Type": "application/json" },
                        })
                        .then((response) => {
                            console.log(response);
                            this.googleAuthenticatorEnabled = true;
                            localStorage.setItem('token', response.data.token);
                        })
                        .catch((err) => {
                            console.log(err);
                            this.enabled2FA = [];
                        });
                    break;
                case "sms":
                    this.smsEnabled = true;
                    break;
                case "fido2":
                    this.fido2Enabled = true;
                    this.submit2FA_FIDO2();
                    break;
            }
        },
        submit2FAGoogleAuth() {
            sendPostRequest(googleAuthVerifyURL, { 'token': localStorage.getItem('token'), 'otp': this.google2FAForm.otp }, success2FAVerifyCallback);
        },
        submit2FASMSVerify() {
            sendPostRequest(smsVerifyURL, { 'token': localStorage.getItem('token'), 'otp': this.sms2FAForm.otp }, success2FAVerifyCallback);
        },
        submit2FASMSSend() {
            sendPostRequest(smsStartURL, this.sms2FAForm.phone,
                (response) => {
                    console.log(response);
                    localStorage.setItem('token', response.data.token);
                },
                (err) => {
                    console.log(err);
                    this.enabled2FA = [];
                })
        },
        submit2FA_FIDO2() {
            sendPostRequest(
                fido2StartURL, { 'token': localStorage.getItem('token') },
                (response) => {
                    console.log("Assertion Options:");
                    console.log(response);
                    localStorage.setItem('token', response.data.token);

                    response.data.assertion.publicKey.challenge = bufferDecode(response.data.assertion.publicKey.challenge);
                    response.data.assertion.publicKey.allowCredentials.forEach(function(listItem) {
                        listItem.id = bufferDecode(listItem.id)
                    });
                    console.log(response.data.assertion);
                    navigator.credentials.get({
                            publicKey: response.data.assertion.publicKey
                        })
                        .then(function(credential) {
                            console.log(credential);
                            app.verifyAssertion(credential);
                        }).catch(function(err) {
                            console.log(err);
                        });
                },
                (err) => {
                    console.log(err);
                    this.enabled2FA = [];
                })
        },
        verifyAssertion(assertedCredential) {
            console.log('calling verify')
            let authData = new Uint8Array(assertedCredential.response.authenticatorData);
            let clientDataJSON = new Uint8Array(assertedCredential.response.clientDataJSON);
            let rawId = new Uint8Array(assertedCredential.rawId);
            let sig = new Uint8Array(assertedCredential.response.signature);
            let userHandle = new Uint8Array(assertedCredential.response.userHandle);

            sendPostRequest(
                fido2VerifyURL,
                JSON.stringify({
                    token: localStorage.getItem('token'),
                    id: assertedCredential.id,
                    rawId: bufferEncode(rawId),
                    type: assertedCredential.type,
                    response: {
                        authenticatorData: bufferEncode(authData),
                        clientDataJSON: bufferEncode(clientDataJSON),
                        signature: bufferEncode(sig),
                        userHandle: bufferEncode(userHandle),
                    },
                }),
                success2FAVerifyCallback
            )
        }
    },
    computed: {
        emailEnabled: function() {
            return this.enabledAuthn.includes('email');
        },
        phoneEnabled: function() {
            return this.enabledAuthn.includes('phone');
        },
        pwbasedEnabled: function() {
            return this.enabledAuthn.includes('password_based');
        },
        googleEnabled: function() {
            return this.enabledAuthn.includes('google');
        },
        facebookEnabled: function() {
            return this.enabledAuthn.includes('facebook');
        },
        vkEnabled: function() {
            return this.enabledAuthn.includes('vk');
        },
        appleEnabled: function() {
            return this.enabledAuthn.includes('apple');
        },
        socialProvidersEnabled: function() {
            return this.enabledAuthn.filter((x) => expectedProviders.includes(x)).length > 0;
        },
        googleAuthenticatorEnabled: {
            get: function() {
                return this.enabled2FA.includes('google_authenticator');
            },
            set: function(newValue) {
                if (newValue) {
                    this.enabled2FA = ['google_authenticator']
                }
            }
        },
        smsEnabled: {
            get: function() {
                return this.enabled2FA.includes('sms');
            },
            set: function(newValue) {
                if (newValue) {
                    this.enabled2FA = ['sms'];
                }
            }
        },
        fido2Enabled: {
            get: function() {
                return this.enabled2FA.includes('fido2');
            },
            set: function(newValue) {
                if (newValue) {
                    this.enabled2FA = ['fido2'];
                }
            }
        },
        show2FAChoice: function() {
            return this.enabled2FA.length > 1;
        },
        show2FA: function() {
            return this.enabled2FA.length == 1;
        }
    },
    mounted() {
        const getEnabledAuthnRequest = axios.get(aureoleURL + getEnabledAuthnURL);
        const getRedirectURLRequest = axios.get(aureoleURL + getRedirectURL);
        const getJWTKeysRequest = axios.get(aureoleURL + getJWTKeysURL);

        axios
            .all([getEnabledAuthnRequest, getRedirectURLRequest, getJWTKeysRequest])
            .then(axios.spread((...responses) => {
                console.log(responses);

                const getEnabledAuthnResp = responses[0];
                const getRedirectURLResp = responses[1];
                const getJWTKeysResp = responses[2];

                this.enabledAuthn = getEnabledAuthnResp.data.authn;
                this.redirectURL = getRedirectURLResp.data['redirect_url'];
                this.jwtStorageKeys = getJWTKeysResp.data.keys;
            }))
            .catch(errors => {
                console.log(errors);
                this.errored = true;
            })
            .finally(() => (this.loading = false));
    },
});