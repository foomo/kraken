FROM scratch

COPY bin/kraken /usr/sbin/kraken

# install ca root certificates
# https://curl.haxx.se/docs/caextract.html
# http://blog.codeship.com/building-minimal-docker-containers-for-go-applications/
ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/certs/ca-certificates.crt

EXPOSE 8888

ENTRYPOINT ["/usr/sbin/kraken"]
