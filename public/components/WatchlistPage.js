import { CollectionPage } from "./CollectionPage.js";

export default class WatchlistPage extends CollectionPage {

    constructor() {
        super(() => app.API.getWatchlist(), "Movie Watchlist")
    }

}
customElements.define("watchlist-page", WatchlistPage);