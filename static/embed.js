(function () {
    // Auto-detect the host from the script's source
    var scriptElement = document.currentScript ||
        document.querySelector('script[src*="embed.js"]');
    var embedHost = scriptElement ?
        scriptElement.src.split('/').slice(0, 3).join('/') :
        '{{HOST}}'; // replace this template with HOST from env in app.go i wont do that

    document.querySelectorAll('[data-url]').forEach(function (el) {
        var targetUrl = el.getAttribute('data-url');
        var theme = el.getAttribute('data-theme') || 'light';

        fetch(embedHost + '/api/embed?url=' +
            encodeURIComponent(targetUrl) + '&theme=' +
            encodeURIComponent(theme))
            .then(res => {
                if (!res.ok) throw new Error('Network response was not ok');
                return res.json();
            })
            .then(data => {
                var imagesHtml = '';
                if (data.images && data.images.length > 0) {
                    if (data.images.length === 1) {
                        imagesHtml = '<img src="' + data.images[0] + '" alt="' + data.title + '" class="embed-img">';
                    } else {
                        imagesHtml = '<div class="embed-grid">';
                        data.images.slice(0, 4).forEach(function (img) {
                            imagesHtml += '<div class="embed-grid-item"><img src="' + img + '" alt="' + data.title + '"></div>';
                        });
                        imagesHtml += '</div>';
                    }
                }

                el.innerHTML = '<div class="embed-container ' + data.theme + '">' +
                    '<div class="embed-card">' +
                    (imagesHtml ? '<div class="embed-media">' + imagesHtml + '</div>' : '') +
                    '<div class="embed-content">' +
                    '<h2 class="embed-title">' + data.title + '</h2>' +
                    '<p class="embed-description">' + data.description + '</p>' +
                    '<a href="' + data.url + '" target="_blank" class="embed-link">Visit website</a>' +
                    '</div>' +
                    '</div>' +
                    '</div>';
            })
            .catch(err => {
                el.innerHTML = '<p style="color:red; font-family:-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;">Embed failed >_</p>';
                console.error(err);
            });
    });
})();
