import { CollectionPage } from "./CollectionPage.js";

export default class FavoritePage extends CollectionPage {

    constructor() {
        super(() => app.API.getFavorites(), "Favorite Movies")
    }

}
customElements.define("favorite-page", FavoritePage);