export const Passkeys = {
    register: async (username) => {
        try {
            // Get registration options with the challenge.
            const response = await fetch('/api/passkey/registration-begin', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    "Authorization": app.Store.jwt ? `Bearer ${app.Store.jwt}` : null
                },
                body: JSON.stringify({ username: username })
            });

            // Check if the options are ok.
            if (!response.ok) {
                const err = await response.json();
                app.showError('Failed to get registration options from server.' + err)
            }

            const options = await response.json();

            // This triggers the browser to display the passkey modal
            // A new public-private-key pair is created.
            const attestationResponse = await SimpleWebAuthnBrowser.startRegistration({ optionsJSON: options.publicKey });

            // Send attestationResponse back to server for verification and storage.
            const verificationResponse = await fetch('/api/passkey/registration-end', {
                method: 'POST',
                credentials: 'same-origin',
                headers: {
                    'Content-Type': 'application/json',
                    "Authorization": app.Store.jwt ? `Bearer ${app.Store.jwt}` : null
                },
                body: JSON.stringify(attestationResponse)
            });

            const msg = await verificationResponse.json();
            if (verificationResponse.ok) {
                app.showError("Your passkey was saved. You can use it next time to login")
            } else {
                app.showError(msg, false);
            }
        } catch (e) {
            app.showError('Error: ' + e.message, false);
        }
    },
    authenticate: async (email) => {
        try {
            // Get login options from your server with the challenge
            const response = await fetch('/api/passkey/authentication-begin', {
                method: 'POST', headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email })
            });
            const options = await response.json();

            // This triggers the browser to display the passkey / WebAuthn modal
            // The challenge has been signed after this.
            const assertionResponse = await SimpleWebAuthnBrowser.startAuthentication({ optionsJSON: options.publicKey });

            // Send assertionResponse back to server for verification.
            const verificationResponse = await fetch('/api/passkey/authentication-end', {
                method: 'POST',
                credentials: 'same-origin',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(assertionResponse)
            });

            const serverResponse = await verificationResponse.json();
            if (serverResponse.success) {
                app.Store.jwt = serverResponse.jwt;
                app.Router.go("/account/")
            } else {
                app.showError(msg, false);
            }
        } catch (e) {
            console.log(e)
            app.showError('We couldn\'t authenticate you using a Passkey', false);
        }
    }
}