# A QR code generator written in Golang
Starts an HTTP server (listening on port 8080) that generates QR codes. Once installed and running (see below), the service accepts the following two parameters:
* `data`: (Required) The (URL encoded) string that should be encoded in the QR code
* `size`: (Optional) The size of the image (default: 250)

E.g. ```http://your-domain.tld:8080/?data=Hello%2C%20world&size=300```
