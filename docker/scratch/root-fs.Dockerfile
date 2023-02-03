
ARG base_image=busybox

FROM ${base_image} as base

RUN set -ex && echo "indraOS - Network Freedom."

RUN set -ex && echo "creating root filesystem" \
    && mkdir -pv /tmp/root-fs \
    && mkdir -pv /tmp/root-fs/etc \
    && mkdir -pv /tmp/root-fs/var \
    && mkdir -pv /tmp/root-fs/bin

RUN set -ex && echo "checking root filesystem" \
    && ls -hal /tmp/root-fs \
    && ls -hal /tmp/root-fs/etc \
    && ls -hal /tmp/root-fs/var \
    && ls -hal /tmp/root-fs/bin

##
## Users and Groups
##

RUN set -ex && echo "adding users and groups" \
    && echo "btcd:*:::::::" >> /etc/shadow \
    && echo "btcd:x:8333:" >> /etc/group \
    && echo "btcd:x:8333:8333:btcd:/var/btcd:/sbin/false" >> /etc/passwd \
    && echo "lnd:*:::::::" >> /etc/shadow \
    && echo "lnd:x:9735:" >> /etc/group \
    && echo "lnd:x:9735:9735:lnd:/var/lnd:/sbin/false" >> /etc/passwd \
    && echo "indra:*:::::::" >> /etc/shadow \
    && echo "indra:x:8337:" >> /etc/group \
    && echo "indra:x:8337:8337:indra:/var/indra:/sbin/false" >> /etc/passwd

RUN set -ex && echo "checking users and groups" \
    && cat /etc/shadow \
    && cat /etc/group \
    && cat /etc/passwd

RUN set -ex && echo "copying users and groups to root filesystem" \
    && cp -p /etc/shadow /tmp/root-fs/etc/shadow \
    && cp -p /etc/group /tmp/root-fs/etc/group \
    && cp -p /etc/passwd /tmp/root-fs/etc/passwd

# DEBUG
RUN set -ex && echo "checking users and groups to root filesystem" \
    && ls -hal /tmp/root-fs/etc \
    && cat /tmp/root-fs/etc/shadow \
    && cat /tmp/root-fs/etc/passwd \
    && cat /tmp/root-fs/etc/group

##
## Configuration and Data directories
##

RUN set -ex && echo "adding and permissioning /etc directories" \
    && mkdir -pv /etc/btcd && chmod 755 /etc/btcd \
    && mkdir -pv /etc/btcd/keys && chmod 750 /etc/btcd/keys && chown btcd:btcd /etc/btcd/keys \
    && mkdir -pv /etc/lnd  && chmod 755 /etc/lnd \
    && mkdir -pv /etc/lnd/keys && chmod 750 /etc/lnd/keys && chown lnd:lnd /etc/lnd/keys \
    && mkdir -pv /etc/lnd/macaroons && chmod 750 /etc/lnd/macaroons && chown lnd:lnd /etc/lnd/macaroons \
    && mkdir -pv /etc/indra && chmod 755 /etc/indra \
    && mkdir -pv /etc/indra/keys && chmod 750 /etc/indra/keys && chown indra:indra /etc/indra/keys

RUN set -ex && echo "adding keys to verify btcd/lnd releases" \
    && wget https://raw.githubusercontent.com/lightningnetwork/lnd/master/scripts/keys/guggero.asc \
    && chmod 555 guggero.asc \
    && mv guggero.asc /etc/btcd/keys/ \
    && wget https://raw.githubusercontent.com/lightningnetwork/lnd/master/scripts/keys/roasbeef.asc \
    && chmod 555 roasbeef.asc \
    && mv roasbeef.asc /etc/lnd/keys/ \
    && wget https://opensource.conformal.com/GIT-GPG-KEY-conformal.txt \
    && chmod 555 GIT-GPG-KEY-conformal.txt \
    && mv GIT-GPG-KEY-conformal.txt /etc/btcd/keys/
#    && wget https://raw.githubusercontent.com/indra-labs/indra/master/keys/greg.stone.asc \
#    && chmod 555 greg.stone.asc \
#    && mv greg.stone.asc /etc/indra/keys/ \
#    && wget https://raw.githubusercontent.com/indra-labs/indra/master/keys/херетик.asc \
#    && chmod 555 херетик.asc \
#    && mv херетик.asc /etc/indra/keys/

ADD ./docker/scratch/defaults/btcd.conf .
ADD ./docker/scratch/defaults/lnd.conf .

RUN set -ex & echo "adding default .conf files" \
    && chmod 755 btcd.conf && mv btcd.conf /etc/btcd/ \
    && chmod 755 lnd.conf && mv lnd.conf /etc/lnd

RUN set -ex && echo "copying /etc directories to root filesystem" \
    && cp -rp /etc/btcd /tmp/root-fs/etc/btcd \
    && cp -rp /etc/lnd /tmp/root-fs/etc/lnd \
    && cp -rp /etc/indra /tmp/root-fs/etc/indra

# DEBUG
RUN set -ex && echo "checking /etc directories on root filesystem" \
    && ls -hal /tmp/root-fs/etc \
    && ls -hal /tmp/root-fs/etc/btcd \
    && ls -hal /tmp/root-fs/etc/btcd/keys \
    && ls -hal /tmp/root-fs/etc/lnd \
    && ls -hal /tmp/root-fs/etc/lnd/keys \
    && ls -hal /tmp/root-fs/etc/indra

RUN set -ex && echo "adding and permissioning /var directories" \
    && mkdir -pv /var/btcd && chmod 750 /var/btcd && chown btcd:btcd /var/btcd \
    && mkdir -pv /var/btcd/.btcd && chmod 750 /var/btcd/.btcd && chown btcd:btcd /var/btcd/.btcd \
    && mkdir -pv /var/lnd && chmod 750 /var/lnd && chown lnd:lnd /var/lnd \
    && mkdir -pv /var/lnd/.lnd && chmod 750 /var/lnd/.lnd && chown lnd:lnd /var/lnd/.lnd \
    && mkdir -pv /var/indra && chmod 750 /var/indra && chown indra:indra /var/indra

RUN set -ex && echo "copying /var directories to root filesystem" \
    && cp -rp /var/btcd /tmp/root-fs/var/btcd \
    && cp -rp /var/lnd /tmp/root-fs/var/lnd \
    && cp -rp /var/indra /tmp/root-fs/var/indra

# DEBUG
RUN set -ex && echo "checking /var directories on root filesystem" \
    && ls -hal /tmp/root-fs/var \
    && ls -hal /tmp/root-fs/var/btcd \
    && ls -hal /tmp/root-fs/var/btcd/.btcd \
    && ls -hal /tmp/root-fs/var/lnd \
    && ls -hal /tmp/root-fs/var/lnd/.lnd \
    && ls -hal /tmp/root-fs/var/indra

WORKDIR /tmp/root-fs

RUN set -ex && echo "building root-fs tarball" \
    && tar -cvzf /tmp/root-fs.tar.gz . \
    && rm -rf /tmp/root-fs \
    && ls -hal /tmp

#RUN set -ex && tar -xzvf /tmp/root-fs.tgz \
#    && ls -hal /tmp \
#    && ls -hal /tmp/root-fs \
#    && ls -hal /tmp/root-fs/etc \
#    && ls -hal /tmp/root-fs/etc/btcd \

##
## Base Image
##

#
# Note: We CANNOT use the scratch container to build the our scratch image.
#
# When using the COPY command between container, docker does not preserve permissions.
# Instead, we will opt for generating a root-fs on the build image and extracting it as a tarball.
#

#FROM scratch
#
## Migrate over users and groups
#COPY --from=base /etc/passwd /etc/passwd
#COPY --from=base /etc/group /etc/group
#
## Configuration
#COPY --from=base /etc/btcd /etc/btcd
#COPY --from=base /etc/lnd /etc/lnd
#COPY --from=base /etc/indra /etc/indra
#
### Data
#COPY --from=base --chown=btcd:btcd /var/btcd /var/btcd
#COPY --from=base --chown=lnd:lnd /var/lnd /var/lnd
#COPY --from=base --chown=indra:indra /var/indra /var/indra
