FROM debian:stable-slim

COPY discodns .
RUN useradd -m -s /usr/sbin/nologin discodns
USER discodns
ENTRYPOINT ["./discodns"]
CMD ["-l", "0.0.0.0", "-p", "9053", "--etcd=http://172.21.0.2:2379", "-m", "0"]
