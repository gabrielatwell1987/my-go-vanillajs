import { routes } from "./Routes.js";

const Router = {
    init: () => {
        document.querySelectorAll("a.navlink").forEach(a => {
            a.addEventListener("click", event => {
                event.preventDefault();
                const href = a.getAttribute("href");
                Router.go(href);
            });
        });
        window.addEventListener("popstate", () => {
            Router.go(location.pathname, false);
        });
        // Process initial URL
        Router.go(location.pathname + location.search);
    },
    go: (route, addToHistory = true) => {
        if (addToHistory) {
            history.pushState(null, "", route);
        }
        const routePath = route.includes('?') ? route.split('?')[0] : route;
        let pageElement = null;
        for (const r of routes) {
            if (typeof r.path === "string" && r.path === routePath) {
                pageElement = new r.component();
                pageElement.loggedIn = r.loggedIn;
            } else if (r.path instanceof RegExp) {
                const match = r.path.exec(route);
                if (match) {
                    const params = match.slice(1);
                    pageElement = new r.component();
                    pageElement.loggedIn = r.loggedIn;

                    pageElement.params = params;
                }
            }
            if (pageElement) {
                // A page was found, we checked if we have access to it.
                if (pageElement.loggedIn && app.Store.loggedIn == false) {
                    app.Router.go("/account/login");
                    return;
                }
                break;
            }
        }

        const oldPage = document.querySelector("main").firstElementChild;
        if (oldPage) oldPage.style.viewTransitionName = "old";
        pageElement.style.viewTransitionName = "new";

        function updatePage() {
            document.querySelector("main").innerHTML = "";
            document.querySelector("main").appendChild(pageElement);
        }

        if (!document.startViewTransition) {
            // We don't do a transition
            updatePage();
        } else {
            // We do a transition
            document.startViewTransition(() => {
                updatePage();
            });
        }

    }
}

export default Router;