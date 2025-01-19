if (document.getElementById('registerButton') != undefined) {
    document.getElementById('registerButton').addEventListener('click', register);
    checkNotWindows10();
}
if (document.getElementById('loginButton') != undefined) {
    document.getElementById('loginButton').addEventListener('click', login);
    fetchLoginInfo();
}

function checkNotWindows10() {
    // Firefoxは未対応
    if (navigator.userAgentData !== undefined) {
        navigator.userAgentData.getHighEntropyValues(["platformVersion"])
            .then(ua => {
                if (navigator.userAgentData.platform === "Windows") {
                    const majorPlatformVersion = parseInt(ua.platformVersion.split('.')[0]);
                    if (majorPlatformVersion >= 13) {
                        console.log("Windows 11 or later");
                    }
                    else if (majorPlatformVersion > 0) {
                        console.log("Windows 10");
                    }
                    else {
                        console.log("Before Windows 10");
                    }
                }
                else {
                    console.log("Not running on Windows");
                }
            });
    }
}

function updateTime() {
    const currenttime = new Date();
    document.getElementById('currenttime').innerText = currenttime.toLocaleString();
}

async function fetchLoginInfo() {
    const url = window.location.href;
    const login_token = url.split("/")[url.split("/").length - 1];
    const lres = await fetch('/api/v1/loginInfo/' + login_token, { method: "GET" });
    const loginInfo = await lres.json();
    console.log(loginInfo.hostname);
    document.getElementById('hostname').innerText = loginInfo.hostname;
    document.getElementById('username').innerText = loginInfo.username;
    document.getElementById('ipaddr').innerText = loginInfo.ipaddr;
    document.getElementById('your_ipaddr').innerText = loginInfo.your_ipaddr;
    const time = new Date(loginInfo.requested_at);
    document.getElementById('requested_at').innerText = time.toLocaleString();
    updateTime();
    setInterval(updateTime, 1000);
}

function showMessage(message, isError = false) {
    const messageElement = document.getElementById('message');
    messageElement.textContent = message;
    messageElement.style.color = isError ? 'red' : 'green';
}

async function register() {
    // Retrieve the username from the input field
    //const username = document.getElementById('username').value;

    try {
        // Get registration options from your server. Here, we also receive the challenge.
        const response = await fetch('/api/v1/passkey/registerStart', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({})
        });

        // Check if the registration options are ok.
        if (!response.ok) {
            const msg = await response.json();
            throw new Error('User already exists or failed to get registration options from server: ' + msg);
        }

        // Convert the registration options to JSON.
        const options = await response.json();

        // This triggers the browser to display the passkey / WebAuthn modal (e.g. Face ID, Touch ID, Windows Hello).
        // A new attestation is created. This also means a new public-private-key pair is created.
        const attestationResponse = await SimpleWebAuthnBrowser.startRegistration(options.publicKey);

        // Send attestationResponse back to server for verification and storage.
        const verificationResponse = await fetch('/api/v1/passkey/registerFinish', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(attestationResponse)
        });


        const msg = await verificationResponse.json();
        if (verificationResponse.ok) {
            showMessage(msg, false);
        } else {
            showMessage(msg, true);
        }
    } catch
    (error) {
        showMessage('Error: ' + error.message, true);
    }
}

async function login() {
    // get path from url
    const url = window.location.href;
    const login_token = url.split("/")[url.split("/").length - 1];

    try {
        // Get login options from your server. Here, we also receive the challenge.
        const response = await fetch('/api/v1/passkey/loginStart/' + login_token, {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({})
        });
        // Check if the login options are ok.
        if (!response.ok) {
            const msg = await response.json();
            throw new Error('Failed to get login options from server: ' + msg);
        }
        // Convert the login options to JSON.
        const options = await response.json();

        // This triggers the browser to display the passkey / WebAuthn modal (e.g. Face ID, Touch ID, Windows Hello).
        // A new assertionResponse is created. This also means that the challenge has been signed.
        const assertionResponse = await SimpleWebAuthnBrowser.startAuthentication(options.publicKey);

        // Send assertionResponse back to server for verification.
        const verificationResponse = await fetch('/api/v1/passkey/loginFinish/' + login_token, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(assertionResponse)
        });

        const msg = await verificationResponse.json();
        if (verificationResponse.ok) {
            showMessage(msg, false);
        } else {
            showMessage(msg, true);
        }
    } catch (error) {
        showMessage('Error: ' + error.message, true);
    }
}