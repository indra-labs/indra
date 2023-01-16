
ARG semver=master

FROM indralabs/btcd:${semver}

ENTRYPOINT ["/bin/btcctl", "--configfile=/etc/btcd/btcd.conf"]
