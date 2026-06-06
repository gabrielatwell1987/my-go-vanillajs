// service-worker.js

const CACHE_NAME = 'my-cache-v2';

// Install event - precache any initial resources if needed
self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(CACHE_NAME)
            .then(() => {
                // Skip waiting to activate immediately
                self.skipWaiting();
            })
    );
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
    event.waitUntil(
        caches.keys().then((cacheNames) => {
            return Promise.all(
                cacheNames.map((cacheName) => {
                    if (cacheName !== CACHE_NAME) {
                        return caches.delete(cacheName);
                    }
                })
            );
        }).then(() => {
            // Take control of clients immediately
            return self.clients.claim();
        })
    );
});

// Fetch event - handle caching strategies
self.addEventListener('fetch', (event) => {
    const requestUrl = new URL(event.request.url);

    if (!requestUrl.protocol.startsWith('http')) {
        return;
    }

    // Handle /api/ requests (network first, cache fallback)
    if (requestUrl.pathname.startsWith('/api/')) {
        event.respondWith(
            fetch(event.request)
                .then((networkResponse) => {
                    // Cache successful network response
                    return caches.open(CACHE_NAME).then((cache) => {
                        cache.put(event.request, networkResponse.clone());
                        return networkResponse;
                    });
                })
                .catch(() => {
                    // If network fails, try cache
                    return caches.match(event.request)
                        .then((cachedResponse) => {
                            return cachedResponse || Promise.reject('No network or cache available');
                        });
                })
        );
    }
    // Handle all other requests (stale-while-revalidate)
    else {
        event.respondWith(
            caches.open(CACHE_NAME).then((cache) => {
                return cache.match(event.request).then((cachedResponse) => {
                    // Start fetching new version in background
                    const fetchPromise = fetch(event.request)
                        .then((networkResponse) => {
                            // Update cache with new response
                            cache.put(event.request, networkResponse.clone());
                            return networkResponse;
                        })
                        .catch((error) => {
                            console.error('Fetch failed:', error);
                        });

                    // Return cached version if available, otherwise wait for network
                    return cachedResponse || fetchPromise;
                });
            })
        );
    }
});