FROM alpine

RUN ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY ./backup2glacier /bin/backup2glacier

ENTRYPOINT ["/bin/backup2glacier"]