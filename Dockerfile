from alpine
entrypoint ["/coreos-netboot-update-trigger"]

env PATH=/bin:/sbin:/coreos/bin:/coreos/sbin
volume /coreos

add coreos-netboot-update-trigger /
