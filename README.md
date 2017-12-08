tiny-shortener is a tiny URL shortener written in Go. Each URL is stored as a regular file for simplicity.

**usage**

1. shorten a URL:
   * send a POST request to `/` with the following two fields:
     1. `url`: the URL to shorten
     2. `key`: the secret key
   * example cURL command: `curl -X POST -d 'url=http://github.com/kamaln7/tiny-shortener' -d 'key=secret_password' http://localhost:5556/`
     * this will create a short URL at `http://localhost:5556/(random string)` that redirects to `http://github.com/kamaln7/tiny-shortener`.
2. look up a URL/serve a redirect:
   * Browse to `http://[path to klein]/[alias]` to access a short URL.

**options**

Only `-key` is required.

```
Usage of tiny-shortener:
  -key string
        upload API Key
  -length int
        code length (default 3)
  -listenAddr string
        listen address (default "127.0.0.1:5556")
  -notFound string
        404 file
  -root string
        root redirect
  -url string
        path to public facing url (default "http://127.0.0.1:5556/")
  -urls string
        path to urls (default "/srv/www/urls/")
```