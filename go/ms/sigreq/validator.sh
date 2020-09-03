routinator \
     --repository-dir=/tmp/routinator/rpki-cache \
     --tal-dir=/tmp/routinator/tals \
     --rrdp-root-cert=/tmp/own_cert/issuer.crt \
     --allow-dubios-hosts \
     validate --asn $1 --prefix $2