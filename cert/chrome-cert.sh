CHROME_SPKI_HASH=`openssl x509 -pubkey -noout -in cert.pem |
    openssl pkey -pubin -outform der |
    openssl dgst -sha256 -binary |
    base64`


echo ".\chrome.exe --allow-insecure-localhost --ignore-certificate-errors-spki-list=$CHROME_SPKI_HASH --origin-to-force-quic-on=localhost:4443 --user-data-dir=.\quic-userdata https://localhost:3000/?url=https://localhost:4443"