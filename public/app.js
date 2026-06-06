import Router from "./services/Router.js";
import Store from "./services/Store.js";

export const API = {
    baseURL: "/api/",
    getTopMovies: async () => {
        return await API.fetch("movies/top/");
    },
    getRandomMovies: async () => {
        return await API.fetch("movies/random/");
    },
    getGenres: async () => {
        return await API.fetch("genres/");
    },
    getMovieById: async (id) => {
        return await API.fetch(`movies/${id}`);
    },
    searchMovies: async (q, order, genre) => {
        return await API.fetch(`movies/search/`, { q, order, genre });
    },
    register: async (name, email, password) => {
        return await API.send("account/register/", { name, email, password })
    },
    login: async (email, password) => {
        return await API.send("account/authenticate/", { email, password })
    },
    getFavorites: async () => {
        return await API.fetch("account/favorites/")
    },
    getWatchlist: async () => {
        return await API.fetch("account/watchlist/")
    },
    saveToCollection: async (movie_id, collection) => {
        return await API.send("account/save-to-collection/", { movie_id, collection })
    },
    send: async (serviceName, data) => {
        try {
            const response = await fetch(API.baseURL + serviceName, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    "Authorization": app.Store.jwt ? `Bearer ${app.Store.jwt}` : null
                },
                body: JSON.stringify(data)
            });
            if (!response.ok) {
                if (response.status === 401) {
                    app.Store.jwt = null;
                    app.Router.go("/account/login");
                    return { success: false, message: "Session expired. Please log in again." };
                }
                return { success: false, message: "Request failed" };
            }
            const result = await response.json();
            return result;
        } catch (e) {
            console.error(e);
            return { success: false, message: "Network error" };
        }
    },
    fetch: async (serviceName, args) => {
        try {
            const queryString = args ? new URLSearchParams(args).toString() : "";
            const response = await fetch(API.baseURL + serviceName + "?" + queryString, {
                headers: {
                    "Authorization": app.Store.jwt ? `Bearer ${app.Store.jwt}` : null
                }
            });
            if (!response.ok) {
                if (response.status === 401) {
                    app.Store.jwt = null;
                    app.Router.go("/account/login");
                    return { success: false, message: "Session expired. Please log in again." };
                }
                return [];
            }
            const result = await response.json();
            return result;
        } catch (e) {
            console.error(e);
            return [];
        }
    }
}

window.addEventListener("DOMContentLoaded", () => {
    app.Router.init()
    // navigator.serviceWorker.register("/sw.js")
})

window.app = {
    API,
    Router,
    Store,

    showError: (message = 'There was an error loading the page', goToHome = true) => {
        document.querySelector("#alert-modal").showModal()
        document.querySelector("#alert-modal p").textContent = message;
        if (goToHome) app.Router.go("/");
        return;
    },
    closeError: () => {
        document.getElementById('alert-modal').close()
    },
    search: (event) => {
        event.preventDefault();
        const keywords = document.querySelector("input[type=search]").value;
        if (keywords.length > 1) {
            app.Router.go(`/movies?q=${keywords}`)
        }
    },
    searchOrderChange: (order) => {
        const urlParams = new URLSearchParams(window.location.search);
        const q = urlParams.get("q");
        const genre = urlParams.get("genre") ?? "";
        app.Router.go(`/movies?q=${q}&order=${order}&genre=${genre}`);
    },
    searchFilterChange: (genre) => {
        const urlParams = new URLSearchParams(window.location.search);
        const q = urlParams.get("q");
        const order = urlParams.get("order") ?? "";
        app.Router.go(`/movies?q=${q}&order=${order}&genre=${genre}`);
    },
    register: async (event) => {
        event.preventDefault();
        let errors = [];
        const name = document.getElementById("register-name").value;
        const email = document.getElementById("register-email").value;
        const password = document.getElementById("register-password").value;
        const passwordConfirm = document.getElementById("register-password-confirm").value;

        if (name.length < 4) errors.push("Enter your complete name");
        if (email.length < 8) errors.push("Enter your complete email");
        if (password.length < 6) errors.push("Enter a password with 6 characters");
        if (password != passwordConfirm) errors.push("Passwords don't match");
        if (errors.length == 0) {
            const response = await API.register(name, email, password);
            if (response.success) {
                app.Store.jwt = response.jwt;
                app.Router.go("/account/")
            } else {
                app.showError(response.message, false);
            }
        } else {
            app.showError(errors.join(". "), false);
        }
    },
    login: async (event) => {
        event.preventDefault();
        let errors = [];
        const email = document.getElementById("login-email").value;
        const password = document.getElementById("login-password").value;

        if (email.length < 8) errors.push("Enter your complete email");
        if (password.length < 6) errors.push("Enter a password with 6 characters");
        if (errors.length == 0) {
            const response = await API.login(email, password);
            if (response.success) {
                app.Store.jwt = response.jwt;
                app.Router.go("/account/")
            } else {
                app.showError(response.message, false);
            }
        } else {
            app.showError(errors.join(". "), false);
        }
    },
    logout: () => {
        app.Store.jwt = null;
        app.Router.go("/");
    },
    getFavorites: async () => {
        try {
            return await API.getFavorites();
        } catch (e) {
            app.Router.go("/account/")
        }
    },
    getWatchlist: async () => {
        try {
            return await API.getWatchlist();
        } catch (e) {
            app.Router.go("/account/")
        }
    },
    saveToCollection: async (movie_id, collection) => {
        if (app.Store.loggedIn) {
            try {
                const response = await API.saveToCollection(movie_id, collection);
                if (response.success) {
                    switch (collection) {
                        case "favorite":
                            app.Router.go("/account/favorites")
                            break;
                        case "watchlist":
                            app.Router.go("/account/watchlist")
                    }
                } else {
                    app.showError("We couldn't save the movie.")
                }
            } catch (e) {
                console.log(e)
            }
        } else {
            app.Router.go("/account/");
        }
    },
    addPasskey: async () => {
        const username = "testuser";
        await Passkeys.register(username);
    },
    loginWithPasskey: async () => {
        const username = document.getElementById("login-email").value;
        if (username.length < 4) {
            app.showError("To use a passkey, enter your email address first")
        } else {
            await Passkeys.authenticate(username);
        }
    }
}